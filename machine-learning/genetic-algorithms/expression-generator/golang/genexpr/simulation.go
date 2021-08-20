package genexpr

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(rand.Int63()))
	},
}

type SimulationParams struct {
	// Number of genes each Chromosome will have
	ChromosomeSize int

	// Max number of digits a term may have before the rest are marked invalid, and truncated.
	// Set to value ≤0 to allow any amount of digits.
	TermMaxDigits int

	// Precision of floating point numbers used in evaluations.
	// Lower numbers may result in precision/accuracy errors leading to false positives and false negatives.
	FloatPrecision uint

	// Maximum possible score a non-exact target can have.
	// Due to how fitness is evaluated — essentially, 1 / abs(result - target) — if
	// a result is only 1 away from the target, its fitness would be 1 / 1 = 1, which
	// is a perfect score. To combat this, we cap the maximum score a non-exact solution can have.
	ImperfectMaxScore float64

	// Multiplier applied to fitness scores of possible solutions which are not
	// whole integers. This serves to discourage answers with decimal parts, which
	// tend to be further from the target than their relative distance on the
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
	// Set to 0 to run without goroutines. Set to -1 to instruct the Simulation to choose
	// an appropriate value, based on the PopulationSize.
	NumGenerationWorkers int
}

type simulationContext struct {
	SimulationParams

	// Enables reuse / allocation amortization of buffers/stacks used within Chromosome.Decode
	decodeStatePool sync.Pool

	bigFloatPool sync.Pool
}

func DefaultSimulationParams() *SimulationParams {
	return &SimulationParams{
		ChromosomeSize: 60,

		TermMaxDigits: -1,
		FloatPrecision: 128,

		ImperfectMaxScore:         0.96,
		NonIntegerScoreMultiplier: 0.2,

		PopulationSize:   50,
		CrossoverRate:    0.8,
		BaseMutationRate: 0.02,

		NumEvaluationWorkers: 0,
		NumGenerationWorkers: -1,
	}
}

type Simulation struct {
	ctx    *simulationContext
	target *big.Float

	iteration  uint
	population Population

	solutions []*Chromosome
}

type Population []*PopulationMember

func (pop Population) Len() int           { return len(pop) }
func (pop Population) Swap(i, j int)      { pop[i], pop[j] = pop[j], pop[i] }
func (pop Population) Less(i, j int) bool { return pop[i] != nil && pop[i].fitness.Cmp(pop[j].fitness) < 0 }

type PopulationMember struct {
	c       *Chromosome
	fitness *big.Float
}

func (member *PopulationMember) Chromosome() *Chromosome {
	return member.c
}

func (member *PopulationMember) Fitness() *big.Float {
	return member.fitness
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
						values:        newStaticBigFloatStack(params.ChromosomeSize, params.FloatPrecision),
					}
				},
			},
			bigFloatPool: sync.Pool{
				New: func() interface{} {
					return NewFloat(params.FloatPrecision)
				},
			},
		},
		iteration: 1,
	}

	if sim.ctx.NumGenerationWorkers == -1 {
		sim.ctx.NumGenerationWorkers = sim.ctx.PopulationSize / 4
	}

	return sim
}

// InitFromInt creates the initial Population and sets the target from an int
func (sim *Simulation) InitFromInt(target int64) {
	sim.Init(big.NewInt(target))
}

// Init creates the initial Population and sets the target
func (sim *Simulation) Init(target *big.Int) {
	sim.target = new(big.Float).SetPrec(sim.ctx.FloatPrecision).SetInt(target)
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
	evaluated, err := chromosome.Decode().Evaluate()
	return &PopulationMember{
		c:       chromosome,
		fitness: sim.calculateFitness(nil, evaluated, err),
	}
}

func (sim *Simulation) calculateFitness(result, evaluated *big.Float, err error) *big.Float {
	if result == nil {
		result = big.NewFloat(-1).SetPrec(sim.ctx.FloatPrecision)
	}

	if err != nil {
		result.SetInt64(0)
		return result
	}

	if sim.target.Cmp(evaluated) == 0 {
		result.SetInt64(1)
		return result
	}

	result.Sub(sim.target, evaluated).Abs(result)
	truncated, _ := result.Int(nil)
	result.SetInt(truncated)

	if result.Cmp(&big.Float{}) == 0 {
		// Avoid division by zero
		result.SetFloat64(sim.ctx.ImperfectMaxScore)
	} else {
		result.Quo(big.NewFloat(sim.ctx.ImperfectMaxScore), result)
	}

	if !evaluated.IsInt() {
		result.Mul(result, big.NewFloat(sim.ctx.NonIntegerScoreMultiplier))
	}

	return result
}

func (sim *Simulation) Population() Population {
	return sim.population
}

func (sim *Simulation) Solutions() []*Chromosome {
	return sim.solutions
}

// Run the Simulation until a solution is found, printing status to the console periodically
func (sim *Simulation) Run() {
	targetString := sim.target.Text('f', 0)
	fmt.Printf("Solving for: %s\n\n", targetString)

	startedAt := time.Now()
	for {
		if sim.Step() {
			fmt.Printf("Iteration %d — SOLVED: %s\n\n", sim.iteration, targetString)
			break
		} else if sim.iteration%100 == 0 {
			fmt.Printf("Iteration %d — solving for: %s\n", sim.iteration, targetString)

			var maxFitness *big.Float = nil
			var fittestChromosome *Chromosome = nil
			for _, simC := range sim.population {
				if maxFitness == nil || simC.fitness.Cmp(maxFitness) > 0 {
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
		if approxFitness, acc := c.fitness.Float64(); approxFitness == 1.0 && acc == big.Exact {
			sim.solutions = append(sim.solutions, c.c)
		}
	}

	return len(sim.solutions) > 0
}

func (sim *Simulation) iteratePopulation() {
	generation := sim.nextGeneration()

	evaluateChromosomes := func(chromosomes []*Chromosome, startIndex int) {
		for i, chromosome := range chromosomes {
			prevFitness := sim.population[startIndex+i].fitness
			evaluated, err := chromosome.Decode().Evaluate()
			sim.population[startIndex+i] = &PopulationMember{
				c:       chromosome,
				fitness: sim.calculateFitness(prevFitness, evaluated, err),
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

		aSim, bSim := sim.selectChromosomePairFromSortedSliceAndRand(chromosomes, rng)
		a, b := aSim.c, bSim.c

		generationMultiplier := 2 - math.Log(float64(sim.iteration%100))/math.Log(100)
		shiftMultiplier := generationMultiplier

		mutationRate := sim.ctx.BaseMutationRate*generationMultiplier - rng.Float64()*sim.ctx.BaseMutationRate*generationMultiplier

		aFitnessApprox, _ := aSim.fitness.Float64()
		bFitnessApprox, _ := bSim.fitness.Float64()
		aMutationRate := mutationRate * (1 - math.Abs(aFitnessApprox) + rng.Float64()*sim.ctx.BaseMutationRate*generationMultiplier)
		bMutationRate := mutationRate * (1 - math.Abs(bFitnessApprox) + rng.Float64()*sim.ctx.BaseMutationRate*generationMultiplier)

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

	totalFitness := sim.ctx.bigFloatPool.Get().(*big.Float).SetFloat64(0)
	pick := sim.ctx.bigFloatPool.Get().(*big.Float).SetFloat64(0)
	current := sim.ctx.bigFloatPool.Get().(*big.Float).SetFloat64(0)
	// Used in summations to avoid temporary allocations
	tmp := sim.ctx.bigFloatPool.Get().(*big.Float).SetFloat64(0)
	defer func() {
		sim.ctx.bigFloatPool.Put(totalFitness)
		sim.ctx.bigFloatPool.Put(pick)
		sim.ctx.bigFloatPool.Put(current)
		sim.ctx.bigFloatPool.Put(tmp)
	}()

	for _, chromosome := range chromosomes {
		tmp.Add(totalFitness, chromosome.fitness)
		tmp, totalFitness = totalFitness, tmp
	}

	chooseChromosome := func(selectionIndex int, chromosomeIndex int) {
		selection[selectionIndex] = chromosomes[chromosomeIndex]
		tmp.Sub(totalFitness, chromosomes[chromosomeIndex].fitness)
		totalFitness, tmp = tmp, totalFitness

		chromosomes.Swap(chromosomeIndex, 0)
		chromosomes = chromosomes[1:]
	}

nextSelection:
	for i := 0; i < n; i++ {
		current.SetFloat64(0)
		tmp.SetFloat64(0)

		pick.Mul(totalFitness, pick.SetFloat64(rng.Float64()))
		for k, chromosome := range chromosomes {
			tmp.Add(current, chromosome.fitness)
			tmp, current = current, tmp

			if current.Cmp(pick) > 0 {
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
	chromosome.genes = chromosome.genes[:len(expression)]

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
