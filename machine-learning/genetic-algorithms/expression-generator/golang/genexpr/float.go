package genexpr

import (
	"math/big"
	"strings"
)

var largeFloat, _, _ = big.ParseFloat(strings.Repeat("1234567890", 15), 10, 256, big.ToZero)

// NewFloat creates a new big.Float with a mantissa slice cap of 9
func NewFloat(prec uint) *big.Float {
	v := new(big.Float).SetPrec(prec)
	v.Add(largeFloat, largeFloat)
	v.SetInt64(0)
	return v
}
