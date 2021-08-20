package genexpr

import (
	"math/rand"
	"testing"
)



func TestSimulation_Run(t *testing.T) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 50
	params.TermMaxDigits = 3

	sim := NewSimulation(params)
	sim.InitFromInt(98765432111)

	sim.Run()
}

func TestSimulation_RunPrec1024(t *testing.T) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 80
	params.TermMaxDigits = 4
	params.FloatPrecision = 1024

	sim := NewSimulation(params)
	sim.InitFromInt(98765432101234567)

	sim.Run()
}

func BenchmarkSimulation_Run(b *testing.B) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 50
	params.TermMaxDigits = 3

	sim := NewSimulation(params)
	sim.InitFromInt(9876543210)

	b.ResetTimer()
	sim.Run()
}

func BenchmarkSimulation_RunMedium(b *testing.B) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 40
	params.TermMaxDigits = 3

	sim := NewSimulation(params)
	sim.InitFromInt(2222222)

	b.ResetTimer()
	sim.Run()
}

func BenchmarkSimulation_RunSmall(b *testing.B) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 20
	params.TermMaxDigits = 3

	sim := NewSimulation(params)
	sim.InitFromInt(1111)

	b.ResetTimer()
	sim.Run()
}
