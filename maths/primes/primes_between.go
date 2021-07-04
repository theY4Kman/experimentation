package main

import (
	"container/heap"
	"flag"
	"fmt"
	"math"
	"sync"
)

func main() {
	var from, to uint64
	var numWorkers int

	flag.Uint64Var(&from, "from", 2, "Find primes from")
	flag.Uint64Var(&to, "to", 100, "Find primes to")
	flag.IntVar(&numWorkers, "workers", 8, "Number of parallel workers to utilize")
	flag.Parse()

	for prime := range FindPrimesConcurrently(from, to, numWorkers) {
		fmt.Println(prime)
	}
}

type primeResult struct {
	n       uint64
	isPrime bool
}

func (res primeResult) String() string {
	var icon string
	if res.isPrime {
		icon = "✔"
	} else {
		icon = "✘"
	}
	return fmt.Sprintf("%d%s", res.n, icon)
}

func FindPrimesConcurrently(from uint64, to uint64, numWorkers int) <-chan uint64 {
	orderedPrimes := make(chan uint64, numWorkers*4)

	if from < 2 {
		from = 2
	}

	go func() {
		primeResults := make(chan primeResult, numWorkers*4)
		go startPrimesWorkers(from, to, numWorkers, primeResults)
		go findPrimesCollector(from, to, primeResults, orderedPrimes)
	}()

	return orderedPrimes
}

func findPrimesCollector(from uint64, to uint64, results <-chan primeResult, orderedPrimes chan<- uint64) {
	findPrimesCollectorHeap(from, to, results, orderedPrimes)
}

type primeResultHeap []primeResult

func (h primeResultHeap) Len() int           { return len(h) }
func (h primeResultHeap) Less(i, j int) bool { return h[i].n < h[j].n }
func (h primeResultHeap) Swap(i, j int) {
	h[i].n, h[j].n, h[i].isPrime, h[j].isPrime = h[j].n, h[i].n, h[j].isPrime, h[i].isPrime
}

func (h *primeResultHeap) Push(x interface{}) {
	*h = append(*h, x.(primeResult))
}

func (h *primeResultHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *primeResultHeap) PeekAt(i int) interface{} {
	old := *h
	n := len(old)
	x := old[n-i]
	return x
}

func (h *primeResultHeap) Peek() interface{} {
	return h.PeekAt(1)
}

func findPrimesCollectorHeap(from uint64, to uint64, results <-chan primeResult, orderedPrimes chan<- uint64) {
	primeHeap := &primeResultHeap{}
	expectedNumbers := primeCandidatesChan(from, to)
	nextExpectedNumber := <-expectedNumbers

	flushHeap := func() {
		if len(*primeHeap) < 1 {
			return
		}

		for ; len(*primeHeap) >= 1 && nextExpectedNumber == (*primeHeap)[0].n; nextExpectedNumber = <-expectedNumbers {
			contigRes := heap.Pop(primeHeap).(primeResult)
			if contigRes.isPrime {
				orderedPrimes <- contigRes.n
			}
		}
	}

	for result := range results {
		heap.Push(primeHeap, result)
		flushHeap()
	}

	for len(*primeHeap) > 0 {
		res := primeHeap.Pop().(primeResult)
		if res.isPrime {
			orderedPrimes <- res.n
		}
	}

	close(orderedPrimes)
}

func findPrimesCollectorMap(from uint64, to uint64, results <-chan primeResult, orderedPrimes chan<- uint64) {
	buffer := make(map[uint64]bool)
	nextNum := from
	for result := range results {
		if result.n == nextNum {
			if result.isPrime {
				orderedPrimes <- result.n
			}
			nextNum++
		} else {
			buffer[result.n] = result.isPrime
		}

		for ; ; nextNum++ {
			if isPrime, exists := buffer[nextNum]; exists {
				delete(buffer, nextNum)
				if isPrime {
					orderedPrimes <- nextNum
				}
			} else {
				break
			}
		}
	}
	close(orderedPrimes)
}

func startPrimesWorkers(from uint64, to uint64, numWorkers int, results chan<- primeResult) {
	done := sync.WaitGroup{}
	numbers := primeCandidatesChan(from, to)

	for workerNum := 0; workerNum < numWorkers; workerNum++ {
		go func() {
			defer done.Done()
			findPrimesWorker(numbers, results)
		}()
		done.Add(1)
	}

	done.Wait()
	close(results)
}

func findPrimesWorker(numbers <-chan uint64, results chan<- primeResult) {
nextFactor:
	for n := range numbers {
		var factor, maxFactor uint64

		if n == 2 {
			goto isPrime
		}

		maxFactor = uint64(math.Sqrt(float64(n)))
		for factor = 2; factor <= maxFactor; factor++ {
			if n%factor == 0 {
				results <- primeResult{
					n:       n,
					isPrime: false,
				}
				continue nextFactor
			}
		}

		isPrime:
		results <- primeResult{
			n:       n,
			isPrime: true,
		}
	}
}

func findPrimesNaive(from uint64, to uint64) <-chan uint64 {
	primes := make(chan uint64)
	go func() {
	nextFactor:
		for n := from; n <= to; n++ {
			maxFactor := uint64(math.Sqrt(float64(n)))

			var factor uint64
			for factor = 2; factor < maxFactor; factor++ {
				if n%factor == 0 {
					continue nextFactor
				}
			}

			primes <- n
		}

		close(primes)
	}()
	return primes
}

// Returns channel yielding all odd numbers between from and to
func primeCandidatesChan(from uint64, to uint64) <-chan uint64 {
	numbers := make(chan uint64, 100)

	if from <= 2 {
		numbers <- 2
	}

	if from % 2 == 0 {
		from++
	}
	if to % 2 == 0 {
		to--
	}

	go func() {
		for n := from; n <= to; n += 2 {
			numbers <- n
		}
		close(numbers)
	}()
	return numbers
}
