package genexpr

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
)

type Chromosome struct {
	genes   []byte
	ctx     *simulationContext
	decoded *DecodeResult
}

type DecodeResult struct {
	// 1:1 mapping of gene to value — this may include invalid chars/expressions
	RawExpression string

	// Parsed gene expression, with invalid chars/expressions omitted
	Expression string

	// A char for each gene indicating if it's valid '+' (included), invalid '-' (omitted), or unknown '?' (also omitted)
	Validity string

	// Cached results of evaluation
	evaluated *big.Float
	evalErr   error
}

type Validity rune

const (
	Valid   Validity = '+'
	Invalid Validity = '-'
	Unknown Validity = '?'
)

type decodeTokenType int8

const (
	tokenTypeUnknown decodeTokenType = iota
	tokenTypeNumber
	tokenTypeOperator
)

type decodeToken struct {
	Type    decodeTokenType
	Len     int
	Chars   []byte
	Indices []int
}

func tokenTypeOfByte(c byte) decodeTokenType {
	switch c {
	case '+', '-', '*', '/':
		return tokenTypeOperator
	default:
		if c >= '0' && c <= '9' {
			return tokenTypeNumber
		} else {
			return tokenTypeUnknown
		}
	}
}

func precedenceOf(op byte) byte {
	if op == '+' || op == '-' {
		return 0
	} else {
		return 1
	}
}

type decodeState struct {
	validityBuf   []byte
	rawExprBuf    []byte
	exprBuf       []byte
	tokenCharsBuf []byte
	indicesBuf    []int

	tokens      []decodeToken
	validTokens []*decodeToken

	ops    *staticByteStack
	values *staticBigFloatStack
}

func (d *decodeState) Reset() {
	d.ops.Reset()
	d.values.Reset()
}

func (c *Chromosome) String() string {
	var buf strings.Builder
	buf.Grow(len(c.genes)*GeneBits + len(c.genes) - 1)

	lastIndex := len(c.genes) - 1
	geneFmt := fmt.Sprintf("%%0%db", GeneBits)
	for i, gene := range c.Genes() {
		buf.WriteString(fmt.Sprintf(geneFmt, gene))
		if i < lastIndex {
			buf.WriteByte(' ')
		}
	}
	return buf.String()
}

func (c *Chromosome) VerboseString() string {
	decoded := c.Decode()
	result, err := decoded.Evaluate()

	strResult := "ERROR"
	if err == nil {
		strResult = fmt.Sprintf("%f", result)
	}

	var explodedBuf strings.Builder
	explodedBuf.Grow(len(c.genes)*GeneBits + len(c.genes) - 1)

	nLSpaces := GeneBits / 2
	nRSpaces := nLSpaces
	//goland:noinspection ALL
	if GeneBits%2 == 0 {
		nRSpaces--
	}

	lSpace := strings.Repeat(" ", nLSpaces)
	rSpace := strings.Repeat(" ", nRSpaces)

	exprIdx := 0
	finalIdx := len(decoded.Validity) - 1
	for i, validity := range decoded.Validity {
		switch validity {
		case rune(Valid):
			explodedBuf.WriteString(strings.Repeat(string(decoded.RawExpression[exprIdx]), GeneBits))
			exprIdx++
		case rune(Invalid):
			explodedBuf.WriteString(lSpace)
			explodedBuf.WriteByte(decoded.RawExpression[exprIdx])
			explodedBuf.WriteString(rSpace)
			exprIdx++
		case rune(Unknown):
			explodedBuf.WriteString(lSpace)
			explodedBuf.WriteByte('?')
			explodedBuf.WriteString(rSpace)
		}

		if i != finalIdx {
			explodedBuf.WriteByte(' ')
		}
	}

	return fmt.Sprintf("%s\n%s\n  %s\n    = %s", c, explodedBuf.String(), decoded.Expression, strResult)
}

func (c *Chromosome) Copy() *Chromosome {
	copied := &Chromosome{
		genes:   make([]byte, len(c.genes)),
		ctx:     c.ctx,
		decoded: nil,
	}
	copy(copied.genes, c.genes)
	return copied
}

func (c *Chromosome) Genes() []byte {
	return c.genes
}

// Mutate creates a new Chromosome with bits randomly flipped based on mutationRate
func (c *Chromosome) Mutate(mutationRate float64) *Chromosome {
	rng := randPool.Get().(*rand.Rand)
	defer randPool.Put(rng)
	return c.MutateWithRand(mutationRate, rng)
}

func (c *Chromosome) MutateWithRand(mutationRate float64, rng *rand.Rand) *Chromosome {
	mutated := c.Copy()

	for i, gene := range mutated.genes {
		for j := GeneBits - 1; j >= 0; j-- {
			if rng.Float64() < mutationRate {
				mask := byte(1) << j
				bit := gene & mask
				if bit > 0 {
					bit = ^bit
					gene &= bit
				} else {
					gene |= mask
				}
			}
		}
		mutated.genes[i] = gene
	}

	return mutated
}

// LRotate shifts a chromosome's genes left by n bits, wrapping around to the right — returns a new Chromosome
func (c *Chromosome) LRotate(n int) *Chromosome {
	nBytes := n / GeneBits
	nBits := n % GeneBits

	shiftedGenes := make([]byte, len(c.genes))
	copy(shiftedGenes, c.genes[nBytes:])
	copy(shiftedGenes[len(c.genes)-nBytes:], c.genes[:nBytes])

	if nBits > 0 {
		// Bits we'll carry over to the last gene's least significant bits
		carry := (GeneMask << (GeneBits - nBits)) & GeneMask & shiftedGenes[0] >> (GeneBits - nBits)

		for i := 0; i < len(shiftedGenes)-1; i++ {
			shiftedGenes[i] = ((shiftedGenes[i] << nBits) & GeneMask) | ((shiftedGenes[i+1] >> (GeneBits - nBits)) & GeneMask)
		}
		shiftedGenes[len(shiftedGenes)-1] = ((shiftedGenes[len(shiftedGenes)-1] << nBits) | carry) & GeneMask
	}

	return &Chromosome{
		genes: shiftedGenes,
		ctx:   c.ctx,
	}
}

func (c *Chromosome) Decode() *DecodeResult {
	if c.decoded != nil {
		return c.decoded
	}

	state := c.ctx.decodeStatePool.Get().(*decodeState)
	defer func() {
		state.Reset()
		c.ctx.decodeStatePool.Put(state)
	}()

	validityBuf := state.validityBuf

	rawExprLen := 0
	rawExprBuf := state.rawExprBuf

	exprLen := 0
	exprBuf := state.exprBuf

	tokenCharsLen := 0
	tokenCharsBuf := state.tokenCharsBuf

	writeByte := func(buf []byte, bufLen *int, c byte) {
		buf[*bufLen] = c
		*bufLen++
	}
	writeBytes := func(buf []byte, bufLen *int, bytes []byte) {
		copy(buf[*bufLen:], bytes)
		*bufLen += len(bytes)
	}

	indicesLen := 0
	indicesBuf := state.indicesBuf

	tokensLen := 0
	tokens := state.tokens

	validTokensLen := 0
	validTokens := state.validTokens

	for i, gene := range c.Genes() {
		if value := geneValuesArray[gene]; value != 0 {
			writeByte(rawExprBuf, &rawExprLen, value)

			toktype := tokenTypeOfByte(value)
			if tokensLen == 0 || toktype == tokenTypeOperator || toktype != tokens[tokensLen-1].Type {
				tokens[tokensLen] = decodeToken{
					Type:    toktype,
					Len:     0,
					Chars:   tokenCharsBuf[tokenCharsLen:],
					Indices: indicesBuf[indicesLen:],
				}
				tokensLen++
			}

			token := &tokens[tokensLen-1]

			token.Chars[token.Len] = value
			token.Indices[token.Len] = i

			token.Len++
			tokenCharsLen++
			indicesLen++

			// NOTE: this validity may be reversed during parsing
			validityBuf[i] = byte(Valid)
		} else {
			validityBuf[i] = byte(Unknown)
		}
	}

	ops := state.ops
	values := state.values
	var evalErr error = nil

	evalOp := func() {
		if evalErr != nil {
			return
		}

		rhs, _ := values.Pop()
		lhs, _ := values.Pop()
		op, _ := ops.Pop()

		var result *big.Float
		switch op {
		case '+':
			result = lhs.Add(lhs, rhs)
		case '-':
			result = lhs.Sub(lhs, rhs)
		case '*':
			result = lhs.Mul(lhs, rhs)
		case '/':
			// Detect and avoid division-by-zero panics
			if rhs.Cmp(&big.Float{}) == 0 {
				evalErr = fmt.Errorf("division by zero")
				values.Checkin(lhs)
				values.Checkin(rhs)
				ops.Reset()
				values.Reset()
				return
			} else {
				result = lhs.Quo(lhs, rhs)
			}
		}

		values.Checkin(rhs)
		values.Push(result)
	}

	for i := 0; i < tokensLen; i++ {
		tok := &tokens[i]
		if tok.Len == 0 {
			continue
		}

		// Cement the length of a token's slices using its bookkeeping Len fields.
		// This is performed here, during parsing, instead of during tokenization
		// to avoid the need for a separate case outside the tokenization loop
		// to handle the last token (assuming this cementing would be performed
		//on the previous token whenever a new token was found)
		tok.Chars = tok.Chars[:tok.Len]
		tok.Indices = tok.Indices[:tok.Len]

		var peek *decodeToken = nil
		var past *decodeToken = nil

		if i+1 < tokensLen {
			peek = &tokens[i+1]
		}
		if validTokensLen > 0 {
			past = validTokens[validTokensLen-1]
		}

		switch tok.Type {
		case tokenTypeNumber:
			// Remove leading zeroes
			for len(tok.Chars) > 1 && tok.Chars[0] == '0' {
				validityBuf[tok.Indices[0]] = byte(Invalid)
				tok.Chars = tok.Chars[1:]
				tok.Indices = tok.Indices[1:]
				tok.Len--
			}

			// Truncate to max digits
			if c.ctx.TermMaxDigits > 0 && len(tok.Chars) > c.ctx.TermMaxDigits {
				for k := c.ctx.TermMaxDigits; k < tok.Len; k++ {
					validityBuf[tok.Indices[k]] = byte(Invalid)
				}
				tok.Len = c.ctx.TermMaxDigits
				tok.Chars = tok.Chars[:tok.Len]
				tok.Indices = tok.Indices[:tok.Len]
			}

			writeBytes(exprBuf, &exprLen, tok.Chars[:tok.Len])
			validTokens[validTokensLen] = tok
			validTokensLen++

			if evalErr == nil {
				// Determine value of token
				num := int64(0)
				for k := 0; k < tok.Len; k++ {
					d := tok.Chars[k]
					num = num*10 + int64(d-'0')
				}

				// Allow prefix - or + for first number
				if values.Size() == 0 && ops.Size() == 1 {
					op, _ := ops.Pop()
					if op == '-' {
						num *= -1
					}
				}

				value := values.Checkout().SetInt64(num)
				err := values.Push(value)
				if err != nil {
					// TODO: curry error (though, stack errors should never happen)
					fmt.Printf("Stack empty when evaluating partial %s (raw %s)\n", string(exprBuf[:exprLen]), string(rawExprBuf[:rawExprLen]))
					panic(err)
				}
			}

		case tokenTypeOperator:
			op := tok.Chars[0]

			// Allow unary + or -, whenever followed by a number
			isValidUnary := (op == '+' || op == '-') && (peek != nil && peek.Type == tokenTypeNumber)

			// Allow other ops if preceded and followed by a number
			isValidBinary := past != nil && past.Type == tokenTypeNumber && peek != nil && peek.Type == tokenTypeNumber

			if isValidUnary || isValidBinary {
				writeByte(exprBuf, &exprLen, op)
				validTokens[validTokensLen] = tok
				validTokensLen++

				if evalErr == nil {
					if ops.Size() > 0 {
						precedence := precedenceOf(op)
						for ops.Size() > 0 {
							topOp, err := ops.Peek()
							if err != nil || precedenceOf(topOp) < precedence {
								break
							}

							evalOp()
						}
					}

					err := ops.Push(op)
					if err != nil {
						// TODO: curry error (though, stack errors should never happen)
						fmt.Printf("Stack empty when evaluating partial %s (raw %s)\n", string(exprBuf[:exprLen]), string(rawExprBuf[:rawExprLen]))
						panic(err)
					}
				}
			} else {
				validityBuf[tok.Indices[0]] = byte(Invalid)
			}
		}
	}

	var evaluated *big.Float = nil
	if evalErr == nil {
		for ops.Size() > 0 {
			evalOp()
		}

		var result *big.Float
		result, evalErr = values.Pop()
		if evalErr == nil {
			evaluated = (&big.Float{}).Copy(result)
			values.Checkin(result)
		}
	}

	c.decoded = &DecodeResult{
		RawExpression: string(rawExprBuf[:rawExprLen]),
		Expression:    string(exprBuf[:exprLen]),
		Validity:      string(validityBuf[:len(c.genes)]),
		evaluated:     evaluated,
		evalErr:       evalErr,
	}
	return c.decoded
}

func (d *DecodeResult) Evaluate() (*big.Float, error) {
	if d.evaluated != nil {
		return d.evaluated, d.evalErr
	} else {
		return nil, d.evalErr
	}
}
