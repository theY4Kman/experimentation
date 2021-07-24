package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"math/rand"
	"testing"
)

func init() {
	RegisterFailHandler(Fail)
}

func Test(t *testing.T) {
	RunSpecs(t, "Genetic Expressions")
}

var _ = Describe("Chromosome", func() {
	DescribeTable("Decode",
		func(geneExpr string, expectedValidity string, expectedDecodedExpr string) {
			params := DefaultSimulationParams()
			params.ChromosomeSize = len(geneExpr)
			params.TermMaxDigits = 3
			sim := NewSimulation(params)

			chromosome, err := sim.EncodeChromosome(geneExpr, true)
			Expect(err).ToNot(HaveOccurred())

			decoded := chromosome.Decode()
			Expect(*decoded).To(MatchFields(IgnoreExtras, Fields{
				"Expression": Equal(expectedDecodedExpr),
				"Validity": Equal(expectedValidity),
			}))
		},
		Entry("1", "1", "+", "1"),
		Entry("1+", "1+", "+-", "1"),
		Entry("1+2", "1+2", "+++", "1+2"),
		Entry("1-2", "1-2", "+++", "1-2"),
		Entry("1*2", "1*2", "+++", "1*2"),
		Entry("1/2", "1/2", "+++", "1/2"),
		Entry("+1+2", "+1+2", "++++", "+1+2"),
		Entry("+1?+2", "+1?+2", "++?++", "+1+2"),
		Entry("+?1?+2", "+?1?+2", "+?+?++", "+1+2"),

		// Test removal of leading zeroes
		Entry("02", "02", "-+", "2"),
		Entry("20", "20", "++", "20"),

		// Freestanding zeroes are alright
		Entry("0", "0", "+", "0"),

		// Test maxDigits
		Entry("1234", "1234", "+++-", "123"),
	)

	DescribeTable("Decode (raw)",
		func(geneString string, expectedValidity string, expectedDecodedExpr string) {
			sim := NewSimulation(DefaultSimulationParams())

			chromosome, err := sim.ChromosomeFromGeneString(geneString)
			Expect(err).ToNot(HaveOccurred())

			decoded := chromosome.Decode()
			Expect(*decoded).To(MatchFields(IgnoreExtras, Fields{
				"Expression": Equal(expectedDecodedExpr),
				"Validity": Equal(expectedValidity),
			}))
		},
		Entry("01011 01100 00101 10100 00011 01011 11101 11111 01000 00101 00110 01100 01011 01111 00011 10010 01100 10110 00011 01110 01110 11001 11111 01111 01010 10111 10001 01100 11000 01000 00011 00110 01111 11101 10111 01000 00100 11011 01101 11001",
			"01011 01100 00101 10100 00011 01011 11101 11111 01000 00101 00110 01100 01011 01111 00011 10010 01100 10110 00011 01110 01110 11001 11111 01111 01010 10111 10001 01100 11000 01000 00011 00110 01111 11101 10111 01000 00100 11011 01101 11001",
			"+++?++??-++++?-?----+++???+??+?+++???++?+?",
			"27*32-*57237388/7-35-94-94"),
	)

	DescribeTable("Evaluate",
		func(geneExpr string, expectedResult float64) {
			sim := NewSimulation(DefaultSimulationParams())

			chromosome, err := sim.EncodeChromosome(geneExpr, true)
			Expect(err).ToNot(HaveOccurred())

			decoded := chromosome.Decode()
			resultBig, err := decoded.Evaluate()
			Expect(err).ToNot(HaveOccurred())

			result, _ := resultBig.Float64()
			Expect(result).To(Equal(expectedResult))
		},

		Entry("1+2", "1+2", 3.0),
		Entry("1-2", "1-2", -1.0),
		Entry("1*2", "1*2", 2.0),
		Entry("1/2", "1/2", 0.5),

		// Ensure prefixes are evaluated correctly
		Entry("+1", "+1", 1.0),
		Entry("-1", "-1", -1.0),

		// Ensure math is correct for some complex expressions
		Entry("8-264*38*265*824+966+92*37*891+285", "8-264*38*265*824+966+92*37*891+285", -2187553297.0),
	)

	DescribeTable("LRotate",
		func(geneString string, nBits int, expectedGeneString string) {
			sim := NewSimulation(DefaultSimulationParams())

			chromosome, err := sim.ChromosomeFromGeneString(geneString)
			Expect(err).ToNot(HaveOccurred())

			shiftedChromosome := chromosome.LRotate(nBits)
			Expect(shiftedChromosome.String()).To(Equal(expectedGeneString))
		},

		Entry("10101 << 1", "10101", 1, "01011"),
		Entry("10101 01010 << 1", "10101 01010", 1, "01010 10101"),
		Entry("11111 00000 << 5", "11111 00000", 5, "00000 11111"),
		Entry("11111 00000 << 6", "11111 00000", 6, "00001 11110"),
	)

	DescribeTable("CrossoverFulcrum",
		func(aGeneString, bGeneString string, fulcrum int, expectedAGeneString, expectedBGeneString string) {
			sim := NewSimulation(DefaultSimulationParams())

			a, err := sim.ChromosomeFromGeneString(aGeneString)
			Expect(err).ToNot(HaveOccurred())

			b, err := sim.ChromosomeFromGeneString(bGeneString)
			Expect(err).ToNot(HaveOccurred())

			newA, newB, err := CrossoverFulcrum(a, b, fulcrum)
			Expect(err).ToNot(HaveOccurred())
			Expect([]string{newA.String(), newB.String()}).To(Equal([]string{expectedAGeneString, expectedBGeneString}))
		},

		Entry("cross(11111, 00000, 3)",
			"11111", "00000", 3,
			"11100", "00011"),
		Entry("cross(11111 11111, 00000 00000, 8)",
			"11111 11111", "00000 00000", 8,
			"11111 11100", "00000 00011"),
	)
})

func BenchmarkSimulation_Run(b *testing.B) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	sim := NewSimulation(DefaultSimulationParams())
	sim.InitFromInt(987654321)
	sim.Run()
}

func BenchmarkSimulation_RunMedium(b *testing.B) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 20
	params.TermMaxDigits = 3

	sim := NewSimulation(params)
	sim.InitFromInt(2222222)
	sim.Run()
}

func BenchmarkSimulation_RunSmall(b *testing.B) {
	rand.Seed(0)  // use static seed for an inkling of repeatability

	params := DefaultSimulationParams()
	params.ChromosomeSize = 20
	params.TermMaxDigits = 3

	sim := NewSimulation(params)
	sim.InitFromInt(1111)
	sim.Run()
}
