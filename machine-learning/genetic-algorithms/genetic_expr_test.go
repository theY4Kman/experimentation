package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"testing"
)

func init() {
	RegisterFailHandler(Fail)
}

func TestChromosome(t *testing.T) {
	RunSpecs(t, "Chromosome")
}

var _ = Describe("Chromosome", func() {
	DescribeTable("Decode",
		func(geneExpr string, expectedValidity string, expectedDecodedExpr string) {
			chromosome, err := EncodeChromosome(geneExpr, DefaultSimulationParams(), true)
			Expect(err).ToNot(HaveOccurred())

			decoded := chromosome.Decode()
			Expect(decoded).To(MatchFields(IgnoreExtras, Fields{
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

	DescribeTable("Evaluate",
		func(geneExpr string, expectedResult float64) {
			chromosome, err := EncodeChromosome(geneExpr, DefaultSimulationParams(), true)
			Expect(err).ToNot(HaveOccurred())

			decoded := chromosome.Decode()
			result, err := decoded.Evaluate()
			Expect(err).ToNot(HaveOccurred())

			Expect(result).To(Equal(expectedResult))
		},

		Entry("1+2", "1+2", 3.0),
		Entry("1-2", "1-2", -1.0),
		Entry("1*2", "1*2", 2.0),
		Entry("1/2", "1/2", 0.5),

		// Ensure prefixes are evaluated correctly
		Entry("+1", "+1", 1.0),
		Entry("-1", "-1", -1.0),
	)
})
