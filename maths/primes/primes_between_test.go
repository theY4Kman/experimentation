package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

var knownPrimes []uint64
var numKnownPrimes uint64
var highestKnownPrime uint64

func init() {
	knownPrimesPath := "known_primes.bin"
	fp, err := os.Open(knownPrimesPath)
	if err != nil {
		panic(fmt.Sprintf("Unable to locate known primes file %s: %s", knownPrimesPath, err))
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	numKnownPrimes, err = binary.ReadUvarint(reader)
	if err != nil {
		panic(fmt.Sprintf("Failed to read length header from known primes file %s: %s", knownPrimesPath, err))
	}

	knownPrimes = make([]uint64, numKnownPrimes)

	for i := uint64(0); i < numKnownPrimes; i++ {
		prime, err := binary.ReadUvarint(reader)
		if err != nil {
			panic(fmt.Sprintf("Failed to read prime #%d of %d from known primes file %s: %s", i+1, numKnownPrimes, knownPrimesPath, err))
		}

		knownPrimes[i] = prime
	}

	highestKnownPrime = knownPrimes[len(knownPrimes)-1]
}

func TestCorrectness(t *testing.T) {
	i := 0
	for prime := range FindPrimesConcurrently(0, highestKnownPrime + 1, 1) {
		if i >= len(knownPrimes) {
			t.Errorf("Too many primes returned")
		}

		knownPrime := knownPrimes[i]
		i++

		if prime != knownPrime {
			t.Errorf("Incorrect prime %d returned, expected: %d", prime, knownPrime)
		}
	}
}

func TestConcurrency(t *testing.T) {
	i := 0
	for prime := range FindPrimesConcurrently(0, highestKnownPrime + 1, 24) {
		if i >= len(knownPrimes) {
			t.Errorf("Too many primes returned")
		}

		knownPrime := knownPrimes[i]
		i++

		if prime != knownPrime {
			t.Errorf("Incorrect prime %d returned, expected: %d", prime, knownPrime)
		}
	}
}
