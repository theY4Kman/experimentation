package main

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
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
	rand.Seed(time.Now().Unix())

	params := DefaultSimulationParams()
	sim := NewSimulation(rand.Intn(99999), params)

	fmt.Printf("Solving for: %d\n\n", sim.solution)

	for ;; {
		if sim.Step() {
			fmt.Printf("Iteration %d — SOLVED\n\n", sim.iteration)
			break
		} else if sim.iteration % 100 == 0 {
			fmt.Printf("Iteration %d — solving for: %d\n", sim.iteration, sim.solution)

			maxFitness := -1.0
			var fittestChromosome *Chromosome = nil
			for _, simC := range sim.population {
				if simC.fitness > maxFitness {
					maxFitness = simC.fitness
					fittestChromosome = simC.c
				}
			}

			fmt.Printf("\n%s\n\n", fittestChromosome.VerboseString())
		}
	}

	for _, chromosome := range sim.Solutions() {
		fmt.Printf("%s\n\n\n", chromosome.VerboseString())
	}
}

func isIntegral(val float64) bool {
	return val == float64(int(val))
}

type SimulationParams struct {
	// Chromosome config
	ChromosomeSize int

	// Chromosome decoding params
	TermMaxDigits int

	// Evaluated expression scoring params
	ImperfectMaxScore         float64
	NonIntegerScoreMultiplier float64

	// Population config
	PopulationSize   int
	CrossoverRate    float64
	BaseMutationRate float64
}

func DefaultSimulationParams() *SimulationParams {
	return &SimulationParams{
		ChromosomeSize: 40,

		TermMaxDigits: 3,

		ImperfectMaxScore:         0.96,
		NonIntegerScoreMultiplier: 0.2,

		PopulationSize:   50,
		CrossoverRate:    0.8,
		BaseMutationRate: 0.01,
	}
}

type Simulation struct {
	params   *SimulationParams
	solution int

	iteration  uint
	population Population

	solutions []*Chromosome
}

type Population []*SimChromosome

func (pop Population) Len() int           { return len(pop) }
func (pop Population) Swap(i, j int)      { pop[i], pop[j] = pop[j], pop[i] }
func (pop Population) Less(i, j int) bool { return pop[i] != nil && pop[i].fitness < pop[j].fitness }

type SimChromosome struct {
	c       *Chromosome
	fitness float64
}

func NewSimulation(solution int, params *SimulationParams) *Simulation {
	sim := &Simulation{
		params:     params,
		solution:   solution,
		iteration:  1,
		population: make([]*SimChromosome, params.PopulationSize),
	}

	for i := range sim.population {
		sim.population[i] = sim.randomChromosome()
	}

	return sim
}

func (sim *Simulation) randomChromosome() *SimChromosome {
	chromosome := RandomChromosome(sim.params)
	return &SimChromosome{
		c:       chromosome,
		fitness: sim.calculateFitness(chromosome.Decode().Evaluate()),
	}
}

func (sim *Simulation) calculateFitness(evaluated float64, err error) float64 {
	if err != nil {
		return 0
	}

	isInteger := isIntegral(evaluated)
	if isInteger && sim.solution == int(evaluated) {
		return 1
	}

	var intBias float64
	if isInteger {
		intBias = sim.params.NonIntegerScoreMultiplier
	} else {
		intBias = 1
	}

	denominator := math.Trunc(math.Abs(float64(sim.solution) - evaluated))
	if denominator == 0 {
		// Avoid division by zero
		return 0
	}

	return sim.params.ImperfectMaxScore * 1 / denominator * intBias
}

func (sim *Simulation) Solutions() []*Chromosome {
	return sim.solutions
}

// Step iterates the population, and returns whether a solution has been found
func (sim *Simulation) Step() bool {
	sim.iteratePopulation()
	for _, c := range sim.population {
		if c.fitness == 1.0 {
			sim.solutions = append(sim.solutions, c.c)
		}
	}

	return len(sim.solutions) > 0
}

func (sim *Simulation) iteratePopulation() {
	generation := sim.nextGeneration()
	for i, chromosome := range generation {
		sim.population[i] = &SimChromosome{
			c:       chromosome,
			fitness: sim.calculateFitness(chromosome.Decode().Evaluate()),
		}
	}

	sim.iteration++
}

// nextGeneration creates a new population from the current one, based on mutation and crossover rates,
// as well as periodic mutation rate increases based on iteration
func (sim *Simulation) nextGeneration() []*Chromosome {
	generation := make([]*Chromosome, sim.params.PopulationSize)

	for i := 0; i < sim.params.PopulationSize; i += 2 {
		aSim, bSim := sim.selectChromosomePair()
		a, b := aSim.c, bSim.c

		generationMultiplier := 2 - math.Log(float64(sim.iteration%100)) / math.Log(100)
		shiftMultiplier := generationMultiplier

		mutationRate := sim.params.BaseMutationRate * generationMultiplier - rand.Float64() * sim.params.BaseMutationRate * generationMultiplier

		aMutationRate := mutationRate * (1 - math.Abs(aSim.fitness) + rand.Float64() * sim.params.BaseMutationRate * generationMultiplier)
		bMutationRate := mutationRate * (1 - math.Abs(bSim.fitness) + rand.Float64() * sim.params.BaseMutationRate * generationMultiplier)

		if rand.Float64() < sim.params.CrossoverRate {
			var err error
			a, b, err = Crossover(a, b)

			if err != nil {
				// TODO: handle gracefully
				panic(err)
			}
		}

		a = a.Mutate(aMutationRate)
		b = b.Mutate(bMutationRate)

		if rand.Float64() < aMutationRate {
			aShift := int(shiftMultiplier * float64(GeneBits))
			a = a.LRotate(aShift)
		}
		if rand.Float64() < bMutationRate {
			bShift := int(shiftMultiplier * float64(GeneBits))
			b = a.LRotate(bShift)
		}

		generation[i] = a
		generation[i+1] = b
	}

	return generation
}

func (sim *Simulation) selectChromosomes(n int) []*SimChromosome {
	selection := make([]*SimChromosome, n)

	chromosomes := make(Population, len(sim.population))
	copy(chromosomes, sim.population)

	chooseChromosome := func(selectionIndex int, chromosomeIndex int) {
		selection[selectionIndex] = chromosomes[chromosomeIndex]
		chromosomes.Swap(chromosomeIndex, 0)
		chromosomes = chromosomes[1:]
	}

nextSelection:
	for i := 0; i < n; i++ {
		sort.Sort(chromosomes)

		totalFitness := 0.0
		for _, chromosome := range chromosomes {
			totalFitness += chromosome.fitness
		}

		pick := rand.Float64() * totalFitness
		current := 0.0
		for k, chromosome := range chromosomes {
			current += math.Abs(chromosome.fitness)
			if current > pick {
				chooseChromosome(i, k)
				continue nextSelection
			}
		}

		// If no selection could be made (which can happen if all fitness values are 0.0),
		// revert to random choice
		chooseChromosome(i, rand.Intn(len(chromosomes)))
	}

	return selection
}

func (sim *Simulation) selectChromosomePair() (*SimChromosome, *SimChromosome) {
	selection := sim.selectChromosomes(2)
	return selection[0], selection[1]
}

type Chromosome struct {
	genes   []byte
	params  *SimulationParams
	decoded *DecodeResult
}

func NewChromosome(params *SimulationParams) *Chromosome {
	return &Chromosome{
		genes:  make([]byte, params.ChromosomeSize),
		params: params,
	}
}

func EncodeExpression(expression string, params *SimulationParams) (*Chromosome, error) {
	return EncodeChromosome(expression, params, false)
}

func EncodeChromosome(expression string, params *SimulationParams, useRandomUnknownGene bool) (*Chromosome, error) {
	if len(expression) > params.ChromosomeSize {
		return nil, fmt.Errorf("expression \"%s\" is longer than params.ChromosomeSize (%d)", expression, params.ChromosomeSize)
	}

	chromosome := NewChromosome(params)
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

func RandomChromosome(params *SimulationParams) *Chromosome {
	chromosome := NewChromosome(params)
	rand.Read(chromosome.genes)
	for i := range chromosome.genes {
		chromosome.genes[i] &= GeneMask
	}
	return chromosome
}

func ChromosomeFromGeneString(geneString string, params *SimulationParams) (*Chromosome, error) {
	geneString = strings.ReplaceAll(geneString, " ", "")
	numGenes := int(math.Ceil(float64(len(geneString)) / float64(GeneBits)))

	genes := make([]byte, numGenes)
	for i := 0; i < numGenes; i++ {
		gene := byte(0)
		for k, c := range geneString[i*5 : (i+1)*5] {
			switch c {
			case '1':
				gene |= 1 << (GeneBits - k - 1)
			case '0':
			default:
				return nil, fmt.Errorf("unrecognized gene string character %c, expected '1' or '0'", c)
			}
		}

		genes[i] = gene
	}

	return &Chromosome{
		genes:  genes,
		params: params,
	}, nil
}

// CrossoverFulcrum creates two new Chromosomes from the provided two,
// with the higher and lower bits swapped at the fulcrum number of bits
func CrossoverFulcrum(a, b *Chromosome, fulcrum int) (*Chromosome, *Chromosome, error) {
	if fulcrum < 0 {
		return nil, nil, fmt.Errorf("fulcrum %d must be a positive number", fulcrum)
	}
	if len(a.genes) != len(b.genes) {
		return nil, nil, fmt.Errorf("expected number of genes in both chromosomes to match (%d != %d)", len(a.genes), len(b.genes))
	}
	if fulcrum >= len(a.genes) * GeneBits {
		return nil, nil, fmt.Errorf("fulcrum %d must not exceed total number of gene bits (%d)", fulcrum, len(a.genes) * GeneBits)
	}

	numGenes := len(a.genes)

	nBytes := fulcrum / GeneBits
	nBits := fulcrum % GeneBits

	nShiftedBytes := 0
	if nBits > 0 {
		nShiftedBytes = 1
	}

	nBytesLeft := nBytes
	nBytesRight := numGenes - nBytes - nShiftedBytes

	newA := a.Copy()
	newB := b.Copy()

	copy(newA.genes, a.genes[:nBytesLeft])
	copy(newA.genes[numGenes - nBytesRight:], b.genes[numGenes - nBytesRight:])

	copy(newB.genes, b.genes[:nBytesLeft])
	copy(newB.genes[numGenes - nBytesRight:], a.genes[numGenes - nBytesRight:])

	if nShiftedBytes > 0 {
		maskRight := GeneMask >> nBits
		maskLeft := ^maskRight & GeneMask

		newA.genes[nBytesLeft] = (a.genes[nBytesLeft] & maskLeft) | (b.genes[nBytesLeft] & maskRight)
		newB.genes[nBytesLeft] = (b.genes[nBytesLeft] & maskLeft) | (a.genes[nBytesLeft] & maskRight)
	}

	return newA, newB, nil
}

// Crossover creates two new Chromosomes from the provided two,
// with the higher and lower bits swapped at a random number of bits
func Crossover(a, b *Chromosome) (*Chromosome, *Chromosome, error) {
	fulcrum := rand.Intn(len(a.genes))
	return CrossoverFulcrum(a, b, fulcrum)
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
	explodedBuf.Grow(len(c.genes) * GeneBits + len(c.genes) - 1)

	nLSpaces := GeneBits / 2
	nRSpaces := nLSpaces
	if GeneBits % 2 == 0 {
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
		genes: make([]byte, len(c.genes)),
		params: c.params,
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
	mutated := c.Copy()

	for i, gene := range mutated.genes {
		for j := GeneBits - 1; j >= 0; j-- {
			if rand.Float64() < mutationRate {
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
			shiftedGenes[i] = ((shiftedGenes[i]<<nBits) & GeneMask) | ((shiftedGenes[i+1]>>(GeneBits-nBits)) & GeneMask)
		}
		shiftedGenes[len(shiftedGenes)-1] = ((shiftedGenes[len(shiftedGenes)-1] << nBits) | carry) & GeneMask
	}

	return &Chromosome{
		genes:  shiftedGenes,
		params: c.params,
	}
}

type DecodeResult struct {
	// 1:1 mapping of gene to value — this may include invalid chars/expressions
	RawExpression string

	// Parsed gene expression, with invalid chars/expressions omitted
	Expression string

	// A char for each gene indicating if it's valid '+' (included), invalid '-' (omitted), or unknown '?' (also omitted)
	Validity string

	// Cached results of evaluation
	evaluated *float64
	evalErr   error
}

func (d *DecodeResult) Evaluate() (float64, error) {
	if d.evaluated != nil {
		return *d.evaluated, d.evalErr
	}

	result, err := gval.Evaluate(d.Expression, nil, ExprLang)
	if err != nil {
		return 0, err
	}

	evaluated := result.(float64)
	d.evaluated = &evaluated
	d.evalErr = err

	return *d.evaluated, d.evalErr
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
	Indices []int
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

func (c *Chromosome) Decode() *DecodeResult {
	if c.decoded != nil {
		return c.decoded
	}

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
				indicesCap := 6  // arbitrarily padded cap for number tokens
				if toktype == tokenTypeOperator {
					indicesCap = 1
				}

				tokens = append(tokens, decodeToken{
					Type:  toktype,
					Chars: strings.Builder{},
					Indices: make([]int, 0, indicesCap),
				})
			}

			token := &tokens[len(tokens)-1]
			token.Chars.WriteByte(byte(value))
			token.Str = token.Chars.String()
			token.Indices = append(token.Indices, i)

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
				validityBuf[tok.Indices[0]] = byte(Invalid)
				tok.Str = tok.Str[1:]
				tok.Indices = tok.Indices[1:]
			}

			// Truncate to max digits
			if len(tok.Str) > c.params.TermMaxDigits {
				for k := c.params.TermMaxDigits; k < len(tok.Str); k++ {
					validityBuf[tok.Indices[k]] = byte(Invalid)
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
				validityBuf[tok.Indices[0]] = byte(Invalid)
			}
		}
	}

	c.decoded = &DecodeResult{
		RawExpression: rawExprBuf.String(),
		Expression:    exprBuf.String(),
		Validity:      string(validityBuf),
	}
	return c.decoded
}
