package main

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"github.com/hashicorp/golang-lru"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

var GeneBits = 5
var GeneMask = byte(255 >> (8 - GeneBits))

var GeneValues = map[byte]byte{
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

var ValueGenes map[byte]byte
var UnknownGenes []byte

var GeneOperators = []byte("+-*/")
var GeneDigits = []byte("01234567890")

var GeneOperatorsSet map[byte]struct{}
var GeneDigitsSet map[byte]struct{}

var exprLang gval.Language

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(rand.Int63()))
	},
}

var expressionResultCache *lru.Cache

func init() {
	ValueGenes = make(map[byte]byte)
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

	GeneOperatorsSet = make(map[byte]struct{})
	for _, op := range GeneOperators {
		GeneOperatorsSet[op] = struct{}{}
	}

	GeneDigitsSet = make(map[byte]struct{})
	for _, digit := range GeneDigits {
		GeneDigitsSet[digit] = struct{}{}
	}


	// Create an arithmetic expressions Language which supports a "+" unary prefix operator
	exprLang = gval.NewLanguage(
		gval.Arithmetic(),
		gval.PrefixOperator("+", func(c context.Context, parameter interface{}) (interface{}, error) {
			p, isFloat := parameter.(float64)
			if !isFloat {
				return nil, fmt.Errorf("expected float, got: %s", parameter)
			}

			return +p, nil
		}),
	)

	var err error
	expressionResultCache, err = lru.New(256)
	if err != nil {
		// TODO: handle more gracefully?
		panic(err)
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	params := DefaultSimulationParams()
	sim := NewSimulation(params)
	sim.Init(rand.Intn(9999999))
	sim.Run()
}

func isIntegral(val float64) bool {
	return val == float64(int(val))
}

type SimulationParams struct {
	// Number of genes each Chromosome will have
	ChromosomeSize int

	// Max number of digits a term may have before the rest are marked invalid, and truncated
	TermMaxDigits int

	// Maximum possible score a non-exact solution can have.
	// Due to how fitness is evaluated — essentially, 1 / abs(result - solution) — if
	// a result is only 1 away from the solution, its fitness would be 1 / 1 = 1, which
	// is a perfect score. To combat this, we cap the maximum score a non-exact solution can have.
	ImperfectMaxScore         float64

	// Multiplier applied to fitness scores of possible solutions which are not
	// whole integers. This serves to discourage answers with decimal parts, which
	// tend to be further from the solution than their relative distance on the
	// number line is.
	NonIntegerScoreMultiplier float64

	// Number of Chromosomes to include in the Simulation.
	// Must be a multiple of 2.
	PopulationSize   int

	// Rate at which two Chromosomes will cross over (swap their low/high bits at a fulcrum)
	// during population iteration.
	CrossoverRate    float64

	// This rate affects how likely mutations (bit flips) occur within Chromosomes,
	// as well as the likelihood each Chromosome will rotate its bits a random amount.
	BaseMutationRate float64

	// Number of workers to utilize when evaluating each Chromosome's expressions
	// each iteration. Set to 0 to run without goroutines.
	NumEvaluationWorkers int

	// Number of workers to utilize when creating new generations of the Population.
	// Set to 0 to run without goroutines.
	NumGenerationWorkers int
}

type simulationContext struct {
	SimulationParams

	// Enables reuse of
	decodeBufPool sync.Pool
}

func DefaultSimulationParams() *SimulationParams {
	return &SimulationParams{
		ChromosomeSize: 40,

		TermMaxDigits: 3,

		ImperfectMaxScore:         0.96,
		NonIntegerScoreMultiplier: 0.2,

		PopulationSize:   50,
		CrossoverRate:    0.8,
		BaseMutationRate: 0.02,

		//XXX///////////////////////////////////////////////////////////////////////////////////////////
		//NumEvaluationWorkers: 0,
		NumEvaluationWorkers: 4,

		//XXX///////////////////////////////////////////////////////////////////////////////////////////
		//NumGenerationWorkers: 0,
		NumGenerationWorkers: 2,
	}
}

type Simulation struct {
	ctx      *simulationContext
	solution int

	iteration  uint
	population Population

	solutions []*Chromosome
}

type Population []*PopulationMember

func (pop Population) Len() int           { return len(pop) }
func (pop Population) Swap(i, j int)      { pop[i], pop[j] = pop[j], pop[i] }
func (pop Population) Less(i, j int) bool { return pop[i] != nil && pop[i].fitness < pop[j].fitness }

type PopulationMember struct {
	c       *Chromosome
	fitness float64
}

func NewSimulation(params *SimulationParams) *Simulation {
	sim := &Simulation{
		ctx: &simulationContext{
			SimulationParams: *params,
			decodeBufPool: sync.Pool{
				New: func() interface{} {
					return make([]byte, params.ChromosomeSize)
				},
			},
		},
		iteration:  1,
	}

	return sim
}

// Init creates the initial Population and sets the target solution
func (sim *Simulation) Init(solution int) {
	sim.solution = solution
	sim.population = make([]*PopulationMember, sim.ctx.PopulationSize)

	for i := range sim.population {
		sim.population[i] = sim.randomMember()
	}
}

func (sim *Simulation) randomMember() *PopulationMember {
	chromosome := sim.RandomChromosome()
	return &PopulationMember{
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
		intBias = 1
	} else {
		intBias = sim.ctx.NonIntegerScoreMultiplier
	}

	denominator := math.Trunc(math.Abs(float64(sim.solution) - evaluated))
	if denominator == 0 {
		// Avoid division by zero
		return 0
	}

	return sim.ctx.ImperfectMaxScore * 1 / denominator * intBias
}

func (sim *Simulation) Solutions() []*Chromosome {
	return sim.solutions
}

// Run the Simulation until a solution is found, printing status to the console periodically
func (sim *Simulation) Run() {
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

	evaluateChromosomes := func(chromosomes []*Chromosome, startIndex int) {
		for i, chromosome := range chromosomes {
			sim.population[startIndex + i] = &PopulationMember{
				c:       chromosome,
				fitness: sim.calculateFitness(chromosome.Decode().Evaluate()),
			}
		}
	}

	if sim.ctx.NumEvaluationWorkers == 0 {
		evaluateChromosomes(generation, 0)
	} else {
		wg := sync.WaitGroup{}

		chunkSize := len(generation) / sim.ctx.NumEvaluationWorkers
		for i := 0; i < len(generation); i += chunkSize {
			start, end := i, i + chunkSize
			if end >= len(generation) {
				end = len(generation) - 1
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				evaluateChromosomes(generation[start:end], start)
			}()
		}
		wg.Wait()
	}

	sim.iteration++
}

// nextGeneration creates a new population from the current one, based on mutation and crossover rates,
// as well as periodic mutation rate increases based on iteration
func (sim *Simulation) nextGeneration() []*Chromosome {
	generation := make([]*Chromosome, sim.ctx.PopulationSize)
	sortedChromosomes := sim.getSortedChromosomes()

	generateChromosomePair := func(i int, rng *rand.Rand) {
		chromosomes := make(Population, len(sortedChromosomes))
		copy(chromosomes, sortedChromosomes)

		aSim, bSim := sim.selectChromosomePairFromSortedSlice(chromosomes)
		a, b := aSim.c, bSim.c

		generationMultiplier := 2 - math.Log(float64(sim.iteration%100)) / math.Log(100)
		shiftMultiplier := generationMultiplier

		mutationRate := sim.ctx.BaseMutationRate * generationMultiplier - rng.Float64() * sim.ctx.BaseMutationRate * generationMultiplier

		aMutationRate := mutationRate * (1 - math.Abs(aSim.fitness) + rng.Float64() * sim.ctx.BaseMutationRate * generationMultiplier)
		bMutationRate := mutationRate * (1 - math.Abs(bSim.fitness) + rng.Float64() * sim.ctx.BaseMutationRate * generationMultiplier)

		if rng.Float64() < sim.ctx.CrossoverRate {
			var err error
			a, b, err = CrossOver(a, b)

			if err != nil {
				// TODO: handle gracefully
				panic(err)
			}
		}

		a = a.MutateWithRand(aMutationRate, rng)
		b = b.MutateWithRand(bMutationRate, rng)

		if rng.Float64() < aMutationRate {
			aShift := int(shiftMultiplier * float64(GeneBits))
			a = a.LRotate(aShift)
		}
		if rng.Float64() < bMutationRate {
			bShift := int(shiftMultiplier * float64(GeneBits))
			b = a.LRotate(bShift)
		}

		generation[i] = a
		generation[i+1] = b
	}

	if sim.ctx.NumGenerationWorkers == 0 {
		rng := randPool.Get().(*rand.Rand)
		defer randPool.Put(rng)

		for i := 0; i < sim.ctx.PopulationSize; i += 2 {
			generateChromosomePair(i, rng)
		}
	} else {
		queue := make(chan int)
		wg := sync.WaitGroup{}

		for i := 0; i < 1; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				rng := randPool.Get().(*rand.Rand)
				defer randPool.Put(rng)

				for i := range queue {
					generateChromosomePair(i, rng)
				}
			}()
		}

		for i := 0; i < sim.ctx.PopulationSize; i += 2 {
			queue <- i
		}
		close(queue)

		wg.Wait()
	}

	return generation
}

func (sim *Simulation) getSortedChromosomes() Population {
	chromosomes := make(Population, len(sim.population))
	copy(chromosomes, sim.population)
	sort.Sort(chromosomes)
	return chromosomes
}

func (sim *Simulation) selectChromosomes(n int) []*PopulationMember {
	return sim.selectChromosomesFromSortedSlice(n, sim.getSortedChromosomes())
}

func (sim *Simulation) selectChromosomesFromSortedSlice(n int, chromosomes Population) []*PopulationMember {
	rng := randPool.Get().(*rand.Rand)
	defer randPool.Put(rng)
	return sim.selectChromosomesFromSortedSliceAndRand(n, chromosomes, rng)
}

func (sim *Simulation) selectChromosomesFromSortedSliceAndRand(n int, chromosomes Population, rng *rand.Rand) []*PopulationMember {
	selection := make([]*PopulationMember, n)

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

		pick := rng.Float64() * totalFitness
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
		chooseChromosome(i, rng.Intn(len(chromosomes)))
	}

	return selection
}

func (sim *Simulation) selectChromosomePair() (*PopulationMember, *PopulationMember) {
	return sim.selectChromosomePairFromSortedSlice(sim.getSortedChromosomes())
}

func (sim *Simulation) selectChromosomePairFromSortedSlice(chromosomes Population) (*PopulationMember, *PopulationMember) {
	rng := randPool.Get().(*rand.Rand)
	defer randPool.Put(rng)
	return sim.selectChromosomePairFromSortedSliceAndRand(chromosomes, rng)
}

func (sim *Simulation) selectChromosomePairFromSortedSliceAndRand(chromosomes Population, rng *rand.Rand) (*PopulationMember, *PopulationMember) {
	selection := sim.selectChromosomesFromSortedSliceAndRand(2, chromosomes, rng)
	return selection[0], selection[1]
}

type Chromosome struct {
	genes   []byte
	ctx  *simulationContext
	decoded *DecodeResult
}

func (sim *Simulation) NewChromosome() *Chromosome {
	return &Chromosome{
		genes:  make([]byte, sim.ctx.ChromosomeSize),
		ctx: sim.ctx,
	}
}

func (sim *Simulation) EncodeExpression(expression string) (*Chromosome, error) {
	return sim.EncodeChromosome(expression, false)
}

func (sim *Simulation) EncodeChromosome(expression string, useRandomUnknownGene bool) (*Chromosome, error) {
	if len(expression) > sim.ctx.ChromosomeSize {
		return nil, fmt.Errorf("expression \"%s\" is longer than ChromosomeSize (%d)", expression, sim.ctx.ChromosomeSize)
	}

	chromosome := sim.NewChromosome()
	for i, value := range expression {
		if gene, isValid := ValueGenes[byte(value)]; isValid {
			chromosome.genes[i] = gene
		} else if useRandomUnknownGene {
			chromosome.genes[i] = UnknownGenes[rand.Intn(len(UnknownGenes))]
		} else {
			return nil, fmt.Errorf("unrecognized gene value %c at position %d", value, i)
		}
	}
	return chromosome, nil
}

func (sim *Simulation) RandomChromosome() *Chromosome {
	chromosome := sim.NewChromosome()
	rand.Read(chromosome.genes)
	for i := range chromosome.genes {
		chromosome.genes[i] &= GeneMask
	}
	return chromosome
}

func (sim *Simulation) ChromosomeFromGeneString(geneString string) (*Chromosome, error) {
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
		ctx: sim.ctx,
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

// CrossOver creates two new Chromosomes from the provided two,
// with the higher and lower bits swapped at a random number of bits
func CrossOver(a, b *Chromosome) (*Chromosome, *Chromosome, error) {
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
		ctx: c.ctx,
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
			shiftedGenes[i] = ((shiftedGenes[i]<<nBits) & GeneMask) | ((shiftedGenes[i+1]>>(GeneBits-nBits)) & GeneMask)
		}
		shiftedGenes[len(shiftedGenes)-1] = ((shiftedGenes[len(shiftedGenes)-1] << nBits) | carry) & GeneMask
	}

	return &Chromosome{
		genes:  shiftedGenes,
		ctx: c.ctx,
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

// LRU-cached results of evaluation
type cachedEvaluation struct {
	evaluated *float64
	evalErr   error
}

func (d *DecodeResult) Evaluate() (float64, error) {
returnDecodeResultCache:
	if d.evaluated != nil {
		return *d.evaluated, d.evalErr
	} else if d.evalErr != nil {
		return 0, d.evalErr
	}

	if cached, isCached := expressionResultCache.Get(d.Expression); isCached {
		cachedRes := cached.(*cachedEvaluation)
		d.evaluated = cachedRes.evaluated
		d.evalErr = cachedRes.evalErr
		goto returnDecodeResultCache
	}

	result, err := exprLang.Evaluate(d.Expression, nil)
	d.evalErr = err

	if err == nil {
		evaluated := result.(float64)
		d.evaluated = &evaluated
	} else {
		d.evaluated = nil
	}

	expressionResultCache.Add(d.Expression, &cachedEvaluation{
		evaluated: d.evaluated,
		evalErr:   d.evalErr,
	})

	goto returnDecodeResultCache
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
	if _, isOperator := GeneOperatorsSet[c]; isOperator {
		return tokenTypeOperator
	} else if _, isDigit := GeneDigitsSet[c]; isDigit {
		return tokenTypeNumber
	} else {
		return tokenTypeUnknown
	}
}

func (c *Chromosome) Decode() *DecodeResult {
	if c.decoded != nil {
		return c.decoded
	}

	validityBuf := c.ctx.decodeBufPool.Get().([]byte)
	defer c.ctx.decodeBufPool.Put(validityBuf)

	rawExprLen := 0
	rawExprBuf := c.ctx.decodeBufPool.Get().([]byte)
	defer c.ctx.decodeBufPool.Put(rawExprBuf)

	exprLen := 0
	exprBuf := c.ctx.decodeBufPool.Get().([]byte)
	defer c.ctx.decodeBufPool.Put(exprBuf)

	tokenCharsLen := 0
	tokenCharsBuf := c.ctx.decodeBufPool.Get().([]byte)
	defer c.ctx.decodeBufPool.Put(tokenCharsBuf)

	writeByte := func(buf []byte, bufLen *int, c byte) {
		buf[*bufLen] = c
		*bufLen++
	}
	writeBytes := func(buf []byte, bufLen *int, bytes []byte) {
		copy(buf[*bufLen:], bytes)
		*bufLen += len(bytes)
	}

	indicesLen := 0
	indicesBuf := make([]int, c.ctx.ChromosomeSize)

	// A guess at how many tokens there will be
	estimatedNumTokens := c.ctx.ChromosomeSize / 3
	tokens := make([]decodeToken, 0, estimatedNumTokens)

	for i, gene := range c.Genes() {
		if value, isKnown := GeneValues[gene]; isKnown {
			writeByte(rawExprBuf, &rawExprLen, value)

			toktype := tokenTypeOfByte(value)
			if len(tokens) == 0 || toktype == tokenTypeOperator || toktype != tokens[len(tokens)-1].Type {
				tokens = append(tokens, decodeToken{
					Type:       toktype,
					Len:        0,
					Chars:      tokenCharsBuf[tokenCharsLen:],
					Indices:    indicesBuf[indicesLen:],
				})
			}

			token := &tokens[len(tokens)-1]

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

	ops := newStaticByteStack(c.ctx.ChromosomeSize)
	values := newStaticFloatStack(c.ctx.ChromosomeSize)

	evalOp := func() {
		lhs, _ := values.Pop()
		rhs, _ := values.Pop()
		op, _ := ops.Pop()

		var result float64
		switch op {
		case '+':
			result = lhs + rhs
		case '-':
			result = lhs - rhs
		case '*':
			result = lhs * rhs
		case '/':
			result = lhs / rhs
		}

		values.Push(result)
	}

	validTokens := make([]*decodeToken, 0, len(tokens))

	for i := range tokens {
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

		if i < len(tokens)-1 {
			peek = &tokens[i+1]
		}
		if len(validTokens) > 0 {
			past = validTokens[len(validTokens)-1]
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
			if len(tok.Chars) > c.ctx.TermMaxDigits {
				for k := c.ctx.TermMaxDigits; k < tok.Len; k++ {
					validityBuf[tok.Indices[k]] = byte(Invalid)
				}
				tok.Len = c.ctx.TermMaxDigits
				tok.Chars = tok.Chars[:tok.Len]
				tok.Indices = tok.Indices[:tok.Len]
			}

			// Determine value of token
			num := 0
			for k := 0; k < tok.Len; k ++ {
				d := tok.Chars[k]
				num = num * 10 + int(d - '0')
			}

			// Allow prefix - or + for first number
			if values.Size() == 0 && ops.Size() == 1 {
				op, _ := ops.Pop()
				if op == '-' {
					num *= -1
				}
			}

			writeBytes(exprBuf, &exprLen, tok.Chars[:tok.Len])
			validTokens = append(validTokens, tok)

			err := values.Push(float64(num))
			if err != nil {
				// TODO: curry error (though, stack errors should never happen)
				panic(err)
			}

		case tokenTypeOperator:
			op := tok.Chars[0]

			// Allow unary + or -, whenever followed by a number
			isValidUnary := (op == '+' || op == '-') && (peek != nil && peek.Type == tokenTypeNumber)

			// Allow other ops if preceded and followed by a number
			isValidBinary := past != nil && past.Type == tokenTypeNumber && peek != nil && peek.Type == tokenTypeNumber

			if isValidUnary || isValidBinary {
				writeByte(exprBuf, &exprLen, op)
				validTokens = append(validTokens, tok)

				if ops.Size() > 0 {
					precedence := precedenceOf(op)
					for ; ops.Size() > 0; {
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
					panic(err)
				}
			} else {
				validityBuf[tok.Indices[0]] = byte(Invalid)
			}
		}
	}

	for ; ops.Size() > 0; {
		evalOp()
	}
	result, err := values.Pop()
	if err != nil {
		// TODO: curry error (though, stack errors should never happen)
		panic(err)
	}

	c.decoded = &DecodeResult{
		RawExpression: string(rawExprBuf[:rawExprLen]),
		Expression:    string(exprBuf[:exprLen]),
		Validity:      string(validityBuf),
		evaluated:     &result,
	}
	return c.decoded
}

type staticFloatStack struct {
	stack []float64
	length int
}

func newStaticFloatStack(length int) *staticFloatStack {
	return &staticFloatStack{
		stack:  make([]float64, length),
		length: 0,
	}
}

func (s *staticFloatStack) Push(v float64) error {
	if s.length >= len(s.stack) {
		return fmt.Errorf("stack has reached maximum capacity (%d)", len(s.stack))
	}

	s.stack[s.length] = v
	s.length++
	return nil
}

func (s *staticFloatStack) Pop() (float64, error) {
	if s.length == 0 {
		return 0, fmt.Errorf("stack is empty")
	}

	s.length--
	return s.stack[s.length], nil
}

func (s *staticFloatStack) Size() int {
	return s.length
}

type staticByteStack struct {
	stack []byte
	length int
}

func newStaticByteStack(length int) *staticByteStack {
	return &staticByteStack{
		stack:  make([]byte, length),
		length: 0,
	}
}

func (s *staticByteStack) Push(b byte) error {
	if s.length + 1 >= len(s.stack) {
		return fmt.Errorf("stack has reached maximum capacity (%d)", len(s.stack))
	}

	s.stack[s.length] = b
	s.length++
	return nil
}

func (s *staticByteStack) Pop() (byte, error) {
	if s.length == 0 {
		return 0, fmt.Errorf("stack is empty")
	}

	s.length--
	return s.stack[s.length], nil
}

func (s *staticByteStack) Peek() (byte, error) {
	if s.length == 0 {
		return 0, fmt.Errorf("stack is empty")
	}

	return s.stack[s.length-1], nil
}

func (s *staticByteStack) Size() int {
	return s.length
}

func precedenceOf(op byte) byte {
	if op == '+' || op == '-' {
		return 0
	} else {
		return 1
	}
}
