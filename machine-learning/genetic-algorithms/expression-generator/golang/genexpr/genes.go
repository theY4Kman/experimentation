package genexpr

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

	// Extra mappings (not optimized for minimal bit-flips required to transition)
	0b00101: '0',
	0b00110: '1',
	0b01000: '2',
	0b01001: '3',
	0b01011: '4',
	0b01100: '5',
	0b01101: '6',
	0b10010: '7',
	0b10011: '8',
	0b10100: '9',
	0b10110: '+',
	0b10111: '-',
	0b11010: '*',
	0b11011: '/',

	// One unassigned gene, for testing purposes
	//0b11100: '?',
}

var geneValuesArray []byte

var ValueGenes map[byte]byte
var UnknownGenes []byte

var GeneOperators = []byte("+-*/")
var GeneDigits = []byte("01234567890")

var GeneOperatorsSet map[byte]struct{}
var GeneDigitsSet map[byte]struct{}

func init() {
	geneValuesArray = make([]byte, 1<<GeneBits)
	ValueGenes = make(map[byte]byte)
	for gene, value := range GeneValues {
		ValueGenes[value] = gene
		geneValuesArray[gene] = value
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
