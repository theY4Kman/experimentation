package main

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"math/rand"
	"strings"
)

var GeneBits = 5
var GeneMask = byte(255 >> (8 - GeneBits))

var GeneValues = map[byte]rune{
	0b0001: '+',
	0b1000: '-',

	0b0101: '*',
	0b1010: '/',

	0b0010: '1',
	0b0100: '9',

	0b0011: '3',
	0b0110: '5',
	0b1100: '7',

	0b1001: '0',

	0b1011: '2',
	0b1101: '4',

	0b0111: '6',
	0b1110: '8',
}

var ValueGenes map[rune]byte
var UnknownGenes []byte

var GeneOperators = "+-*/"
var GeneDigits = "01234567890"

var ExprLang gval.Language

func init() {
	ValueGenes = make(map[rune]byte)
	for gene, value := range GeneValues {
		ValueGenes[value] = gene
	}

	UnknownGenes = make([]byte, 1<<GeneBits-len(GeneValues))
	for i, n := 0, byte(0); n < 1<<GeneBits; n++ {
		if _, isValidGene := GeneValues[n]; !isValidGene {
			UnknownGenes[i] = n
			i++
		}
	}

	// Create an arithmetic expressions Language which supports a "+" unary prefix operator
	ExprLang = gval.NewLanguage(
		gval.Arithmetic(),
		gval.PrefixOperator("+", func(c context.Context, parameter interface{}) (interface{}, error) {
			p, isFloat := parameter.(float64)
			if !isFloat {
				return nil, fmt.Errorf("expected float, got: %s", parameter)
			}

			return +p, nil
		}),
	)
}

func main() {
	params := DefaultSimulationParams()

	for i := 0; i < 20; i++ {
		chromosome := RandomChromosome(30, params)

		decoded := chromosome.Decode()
		result, err := decoded.Evaluate()
		var strResult string
		if err == nil {
			strResult = fmt.Sprintf("%f", result)
		} else {
			strResult = "ERROR"
		}

		fmt.Printf("%s\n  %s\n    = %s\n\n\n", chromosome, decoded.Expression, strResult)
	}
}

type SimulationParams struct {
	// Chromosome config
	ChromosomeSize int

	// Chromosome decoding params
	TermMaxDigits int

	// Evaluated expression scoring params
	ImperfectMaxScore         float32
	NonIntegerScoreMultiplier float32

	// Population config
	PopulationSize   int
	CrossoverRate    float32
	BaseMutationRate float32
}

func DefaultSimulationParams() *SimulationParams {
	return &SimulationParams{
		ChromosomeSize: 40,

		TermMaxDigits: 3,

		ImperfectMaxScore:         0.96,
		NonIntegerScoreMultiplier: 0.2,

		PopulationSize:   30,
		CrossoverRate:    0.8,
		BaseMutationRate: 0.01,
	}
}

type Chromosome struct {
	genes  []byte
	params *SimulationParams
}

func NewChromosome(numGenes uint8, params *SimulationParams) *Chromosome {
	return &Chromosome{
		genes:  make([]byte, numGenes),
		params: params,
	}
}

func EncodeExpression(expression string, params *SimulationParams) (*Chromosome, error) {
	return EncodeChromosome(expression, params, false)
}

func EncodeChromosome(expression string, params *SimulationParams, useRandomUnknownGene bool) (*Chromosome, error) {
	chromosome := NewChromosome(uint8(len(expression)), params)
	for i, value := range expression {
		if gene, isValid := ValueGenes[value]; isValid {
			chromosome.genes[i] = gene
		} else if useRandomUnknownGene {
			chromosome.genes[i] = UnknownGenes[rand.Intn(len(UnknownGenes))]
		} else {
			return nil, fmt.Errorf("unrecognized gene value %c at position %d", value, i)
		}
	}
	return chromosome, nil
}

func RandomChromosome(numGenes uint8, params *SimulationParams) *Chromosome {
	chromosome := NewChromosome(numGenes, params)
	rand.Read(chromosome.genes)
	for i := range chromosome.genes {
		chromosome.genes[i] &= GeneMask
	}
	return chromosome
}

func (c *Chromosome) String() string {
	var buf strings.Builder
	buf.Grow(len(c.genes)*GeneBits + len(c.genes) - 1)

	lastIndex := len(c.genes) - 1
	geneFmt := fmt.Sprintf("%%0%db", GeneBits)
	for i, gene := range c.Genes() {
		buf.WriteString(fmt.Sprintf(geneFmt, gene))
		if i < lastIndex {
			buf.WriteRune(' ')
		}
	}
	return buf.String()
}

func (c *Chromosome) Genes() []byte {
	return c.genes
}

type DecodeResult struct {
	// 1:1 mapping of gene to value — this may include invalid chars/expressions
	RawExpression string

	// Parsed gene expression, with invalid chars/expressions omitted
	Expression string

	// A char for each gene indicating if it's valid '+' (included), invalid '-' (omitted), or unknown '?' (also omitted)
	Validity string
}

func (d DecodeResult) Evaluate() (float64, error) {
	result, err := gval.Evaluate(d.Expression, nil, ExprLang)
	if err != nil {
		return 0, err
	}

	return result.(float64), err
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
	Type  decodeTokenType
	Chars strings.Builder
	Str   string
	Index int
}

func tokenTypeOfRune(rune rune) decodeTokenType {
	if strings.ContainsRune(GeneOperators, rune) {
		return tokenTypeOperator
	} else if strings.ContainsRune(GeneDigits, rune) {
		return tokenTypeNumber
	} else {
		return tokenTypeUnknown
	}
}

func (c *Chromosome) Decode() DecodeResult {
	var rawExprBuf, exprBuf strings.Builder
	validityBuf := make([]byte, 0, len(c.genes))

	rawExprBuf.Grow(len(c.genes))
	exprBuf.Grow(len(c.genes))

	var tokens []decodeToken

	for i, gene := range c.Genes() {
		if value, isKnown := GeneValues[gene]; isKnown {
			rawExprBuf.WriteRune(value)

			toktype := tokenTypeOfRune(value)
			if len(tokens) == 0 || toktype == tokenTypeOperator || toktype != tokens[len(tokens)-1].Type {
				tokens = append(tokens, decodeToken{
					Type:  toktype,
					Chars: strings.Builder{},
					Index: i,
				})
			}

			token := &tokens[len(tokens)-1]
			token.Chars.WriteByte(byte(value))
			token.Str = token.Chars.String()

			// NOTE: this validity may be reversed during parsing
			validityBuf = append(validityBuf, byte(Valid))
		} else {
			validityBuf = append(validityBuf, byte(Unknown))
		}
	}

	validTokens := make([]*decodeToken, 0, len(tokens))

	for i := range tokens {
		tok := &tokens[i]
		if len(tok.Str) == 0 {
			continue
		}

		var peek *decodeToken = nil
		var past *decodeToken = nil

		if i < len(tokens)-1 {
			peek = &tokens[i+1]
		}
		if len(validTokens) > 0 {
			past = validTokens[len(validTokens)-1]
		}

		switch tok.Type {
		case tokenTypeNumber:
			// Remove leading zeroes
			for len(tok.Str) > 1 && tok.Str[0] == '0' {
				validityBuf[tok.Index] = byte(Invalid)
				tok.Str = tok.Str[1:]
				tok.Index++
			}

			// Truncate to max digits
			if len(tok.Str) > c.params.TermMaxDigits {
				for k := c.params.TermMaxDigits; k < len(tok.Str); k++ {
					validityBuf[tok.Index+k] = byte(Invalid)
				}
				tok.Str = tok.Str[:c.params.TermMaxDigits]
			}

			exprBuf.WriteString(tok.Str)
			validTokens = append(validTokens, tok)

		case tokenTypeOperator:
			op := tok.Str[0]

			// Allow unary + or -, whenever followed by a number
			isValidUnary := (op == '+' || op == '-') && (peek != nil && peek.Type == tokenTypeNumber)

			// Allow other ops if preceded and followed by a number
			isValidBinary := past != nil && past.Type == tokenTypeNumber && peek != nil && peek.Type == tokenTypeNumber

			if isValidUnary || isValidBinary {
				exprBuf.WriteByte(op)
				validTokens = append(validTokens, tok)
			} else {
				validityBuf[tok.Index] = byte(Invalid)
			}
		}
	}

	return DecodeResult{
		RawExpression: rawExprBuf.String(),
		Expression:    exprBuf.String(),
		Validity:      string(validityBuf),
	}
}
