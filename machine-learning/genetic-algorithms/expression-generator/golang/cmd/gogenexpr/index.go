package main

import (
	"flag"
	"fmt"
	"github.com/they4kman/experimentation/machine-learning/genetic-algorithms/expression-generator/golang/genexpr"
	"math/big"
	"math/rand"
	"time"
)

type BigIntValue struct {
	bigInt *big.Int
}

func (v BigIntValue) String() string {
	if v.bigInt != nil {
		return v.bigInt.String()
	} else {
		return ""
	}
}

func (v BigIntValue) Set(s string) error {
	if _, ok := v.bigInt.SetString(s, 10); !ok {
		return fmt.Errorf("malformed integer %q", s)
	}

	return nil
}

func main() {
	rand.Seed(time.Now().Unix())

	useGui := false

	target := new(big.Int)
	wasTargetProvided := false

	maxRandomTarget := new(big.Int)
	wasMaxRandomTargetProvided := false

	flag.BoolVar(&useGui, "gui", false, "Use the GUI")
	flag.Var(&BigIntValue{target}, "target", "Solution to search for. A random target will be selected if not provided")
	flag.Var(&BigIntValue{maxRandomTarget}, "max-random-target", "If no explicit target is provided, this dictates the maximum value of the randomly-selected target (default of 2**min(chromosome size, precision/2) is used)")

	params := genexpr.DefaultSimulationParams()
	flag.IntVar(&params.ChromosomeSize, "chromosome-size", params.ChromosomeSize, "Number of genes in each chromosome")
	flag.IntVar(&params.TermMaxDigits, "max-digits", params.TermMaxDigits, "Maximum number of digits allowed in a number term")
	flag.UintVar(&params.FloatPrecision, "precision", params.FloatPrecision, "Precision to use for floating point numbers during evaluation")
	flag.Float64Var(&params.ImperfectMaxScore, "imperfect-max-score", params.ImperfectMaxScore, "Maximum possible score allowed for an imperfect target")
	flag.Float64Var(&params.NonIntegerScoreMultiplier, "non-integer-score-multiplier", params.NonIntegerScoreMultiplier, "Multiplier applied to scores of non-integer solutions, usually as a negative bias")
	flag.IntVar(&params.PopulationSize, "population-size", params.PopulationSize, "Number of chromosomes in the population")
	flag.Float64Var(&params.CrossoverRate, "crossover-rate", params.CrossoverRate, "Rate at which two chromosomes will cross over (have their low/high bits swapped at a random fulcrum)")
	flag.Float64Var(&params.BaseMutationRate, "base-mutation-rate", params.BaseMutationRate, "Affects the likelihood and rate of mutations and rotations")
	flag.IntVar(&params.NumGenerationWorkers, "num-generation-workers", params.NumGenerationWorkers, "Number of goroutines to utilize when creating new population generations. Set to 0 to disable concurrency.")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "target":
			wasTargetProvided = true
		case "max-random-target":
			wasMaxRandomTargetProvided = true
		}
	})

	getTarget := func() *big.Int {
		if !wasTargetProvided {
			if !wasMaxRandomTargetProvided {
				baseTwoMaxExponent := params.FloatPrecision / 2
				if baseTwoMaxExponent > uint(params.ChromosomeSize) {
					baseTwoMaxExponent = uint(params.ChromosomeSize)
				}

				maxRandomTarget.SetInt64(2).Exp(maxRandomTarget, big.NewInt(int64(baseTwoMaxExponent)), nil)
			}

			rng := rand.New(rand.NewSource(time.Now().Unix()))
			target.Rand(rng, maxRandomTarget)
		}

		return target
	}

	sim := genexpr.NewSimulation(params)
	sim.Init(getTarget())

	if useGui {
		genexpr.GuiMain(sim, getTarget)
	} else {
		sim.Run()
	}
}
