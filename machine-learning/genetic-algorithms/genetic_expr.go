package main

import (
	"fmt"
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
	0b00100: '+',
	0b01110: '-',

	0b01010: '*',
	0b10101: '/',

	0b00000: '0',
	0b00001: '1',
	0b00010: '2',
	0b00011: '2',
	0b00111: '3',
	0b01111: '4',
	0b11111: '5',
	0b11110: '6',
	0b11101: '7',
	0b11000: '8',
	0b11001: '8',
	0b10001: '9',
	0b10000: '0',

	// Currently unassigned genes
	//0b00101: '?',
	//0b00110: '?',
	//0b01000: '?',
	//0b01001: '?',
	//0b01011: '?',
	//0b01100: '?',
	//0b01101: '?',
	//0b10010: '?',
	//0b10011: '?',
	//0b10100: '?',
	//0b10110: '?',
	//0b10111: '?',
	//0b11010: '?',
	//0b11011: '?',
	//0b11100: '?',
}

var ValueGenes map[byte]byte
var UnknownGenes []byte

var GeneOperators = []byte("+-*/")
var GeneDigits = []byte("01234567890")

var GeneOperatorsSet map[byte]struct{}
var GeneDigitsSet map[byte]struct{}

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(rand.Int63()))
	},
}

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
	ImperfectMaxScore float64

	// Multiplier applied to fitness scores of possible solutions which are not
	// whole integers. This serves to discourage answers with decimal parts, which
	// tend to be further from the solution than their relative distance on the
	// number line is.
	NonIntegerScoreMultiplier float64

	// Number of Chromosomes to include in the Simulation.
	// Must be a multiple of 2.
	PopulationSize int

	// Rate at which two Chromosomes will cross over (swap their low/high bits at a fulcrum)
	// during population iteration.
	CrossoverRate float64

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

	// Enables reuse / allocation amortization of buffers/stacks used within Chromosome.Decode
	decodeStatePool sync.Pool
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
	values *staticFloatStack
}

func (d *decodeState) Reset() {
	d.ops.Reset()
	d.values.Reset()
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
		NumGenerationWorkers: 4,
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
			decodeStatePool: sync.Pool{
				New: func() interface{} {
					numByteBufs := 4
					byteBuf := make([]byte, numByteBufs*params.ChromosomeSize)

					getBuf := func(i int) []byte {
						return byteBuf[i*params.ChromosomeSize : (i+1)*params.ChromosomeSize]
					}

					return &decodeState{
						validityBuf:   getBuf(0),
						rawExprBuf:    getBuf(1),
						exprBuf:       getBuf(2),
						tokenCharsBuf: getBuf(3),
						indicesBuf:    make([]int, params.ChromosomeSize),
						tokens:        make([]decodeToken, params.ChromosomeSize),
						validTokens:   make([]*decodeToken, params.ChromosomeSize),
						ops:           newStaticByteStack(params.ChromosomeSize),
						values:        newStaticFloatStack(params.ChromosomeSize),
					}
				},
			},
		},
		iteration: 1,
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

func (sim *Simulation) Iteration() uint {
	return sim.iteration
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

	startedAt := time.Now()
	for {
		if sim.Step() {
			fmt.Printf("Iteration %d — SOLVED\n\n", sim.iteration)
			break
		} else if sim.iteration%100 == 0 {
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
	elapsed := time.Since(startedAt)

	for _, chromosome := range sim.Solutions() {
		fmt.Printf("%s\n\n\n", chromosome.VerboseString())
	}

	fmt.Printf("Elapsed time: %s\n", elapsed)
	avgIterationTime := time.Duration(uint(elapsed) / sim.iteration)
	fmt.Printf("Avg iteration runtime: %s\n", avgIterationTime)

	fmt.Println()
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
			sim.population[startIndex+i] = &PopulationMember{
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
			start, end := i, i+chunkSize
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

		generationMultiplier := 2 - math.Log(float64(sim.iteration%100))/math.Log(100)
		shiftMultiplier := generationMultiplier

		mutationRate := sim.ctx.BaseMutationRate*generationMultiplier - rng.Float64()*sim.ctx.BaseMutationRate*generationMultiplier

		aMutationRate := mutationRate * (1 - math.Abs(aSim.fitness) + rng.Float64()*sim.ctx.BaseMutationRate*generationMultiplier)
		bMutationRate := mutationRate * (1 - math.Abs(bSim.fitness) + rng.Float64()*sim.ctx.BaseMutationRate*generationMultiplier)

		if rng.Float64() < sim.ctx.CrossoverRate {
			var err error
			a, b, err = CrossOverWithRand(a, b, rng)

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

// CrossOver creates two new Chromosomes from the provided two,
// with the higher and lower bits swapped at a random number of bits
func (sim *Simulation) CrossOver(a, b *Chromosome) (*Chromosome, *Chromosome, error) {
	rng := randPool.Get().(*rand.Rand)
	defer randPool.Put(rng)
	return CrossOverWithRand(a, b, rng)
}

func CrossOverWithRand(a, b *Chromosome, rng *rand.Rand) (*Chromosome, *Chromosome, error) {
	fulcrum := rng.Intn(len(a.genes))
	return CrossoverFulcrum(a, b, fulcrum)
}

type Chromosome struct {
	genes   []byte
	ctx     *simulationContext
	decoded *DecodeResult
}

func (sim *Simulation) NewChromosome() *Chromosome {
	return &Chromosome{
		genes: make([]byte, sim.ctx.ChromosomeSize),
		ctx:   sim.ctx,
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
		genes: genes,
		ctx:   sim.ctx,
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
	if fulcrum >= len(a.genes)*GeneBits {
		return nil, nil, fmt.Errorf("fulcrum %d must not exceed total number of gene bits (%d)", fulcrum, len(a.genes)*GeneBits)
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
	copy(newA.genes[numGenes-nBytesRight:], b.genes[numGenes-nBytesRight:])

	copy(newB.genes, b.genes[:nBytesLeft])
	copy(newB.genes[numGenes-nBytesRight:], a.genes[numGenes-nBytesRight:])

	if nShiftedBytes > 0 {
		maskRight := GeneMask >> nBits
		maskLeft := ^maskRight & GeneMask

		newA.genes[nBytesLeft] = (a.genes[nBytesLeft] & maskLeft) | (b.genes[nBytesLeft] & maskRight)
		newB.genes[nBytesLeft] = (b.genes[nBytesLeft] & maskLeft) | (a.genes[nBytesLeft] & maskRight)
	}

	return newA, newB, nil
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
	} else {
		return 0, d.evalErr
	}
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
		if value, isKnown := GeneValues[gene]; isKnown {
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
			for k := 0; k < tok.Len; k++ {
				d := tok.Chars[k]
				num = num*10 + int(d-'0')
			}

			// Allow prefix - or + for first number
			if values.Size() == 0 && ops.Size() == 1 {
				op, _ := ops.Pop()
				if op == '-' {
					num *= -1
				}
			}

			writeBytes(exprBuf, &exprLen, tok.Chars[:tok.Len])
			validTokens[validTokensLen] = tok
			validTokensLen++

			err := values.Push(float64(num))
			if err != nil {
				// TODO: curry error (though, stack errors should never happen)
				fmt.Printf("Stack empty when evaluating partial %s (raw %s)\n", string(exprBuf[:exprLen]), string(rawExprBuf[:rawExprLen]))
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
				validTokens[validTokensLen] = tok
				validTokensLen++

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
			} else {
				validityBuf[tok.Indices[0]] = byte(Invalid)
			}
		}
	}

	for ops.Size() > 0 {
		evalOp()
	}

	var evaluated *float64 = nil
	result, evalErr := values.Pop()
	if evalErr == nil {
		evaluated = &result
	}

	c.decoded = &DecodeResult{
		RawExpression: string(rawExprBuf[:rawExprLen]),
		Expression:    string(exprBuf[:exprLen]),
		Validity:      string(validityBuf),
		evaluated:     evaluated,
		evalErr:       evalErr,
	}
	return c.decoded
}

type staticFloatStack struct {
	stack  []float64
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

func (s *staticFloatStack) Reset() {
	s.length = 0
}

type staticByteStack struct {
	stack  []byte
	length int
}

func newStaticByteStack(length int) *staticByteStack {
	return &staticByteStack{
		stack:  make([]byte, length),
		length: 0,
	}
}

func (s *staticByteStack) Push(b byte) error {
	if s.length+1 >= len(s.stack) {
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

func (s *staticByteStack) Reset() {
	s.length = 0
}

func precedenceOf(op byte) byte {
	if op == '+' || op == '-' {
		return 0
	} else {
		return 1
	}
}
