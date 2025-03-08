package main

import "testing"

var divergentOffset = complex(0, 0)
var convergentOffset = complex(-0.150321, -0.141980)

//func TestMandelbrotOptimizedEscape3(t *testing.T) {
//	//c := complex(0, 0)
//	c := complex(-0.757266, 0.102038)
//	got := mandelbrotOptimizedEscape3(c)
//	want := 0
//	if got != want {
//		t.Errorf("got %v, want %v", got, want)
//	}
//}

func BenchmarkMandelbrotNaiveComplexCond_Convergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotNaiveComplexCond(convergentOffset)
	}
}

func BenchmarkMandelbrotNaiveFloatCond_Convergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotNaiveFloatCond(convergentOffset)
	}
}

func BenchmarkMandelbrotOptimizedEscape1_Convergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotOptimizedEscape1(convergentOffset)
	}
}

func BenchmarkMandelbrotOptimizedEscape2_Convergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotOptimizedEscape2(convergentOffset)
	}
}

func BenchmarkMandelbrotOptimizedEscape3_Convergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotOptimizedEscape3(convergentOffset)
	}
}

func BenchmarkMandelbrotNaiveComplexCond_Divergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotNaiveComplexCond(divergentOffset)
	}
}

func BenchmarkMandelbrotNaiveFloatCond_Divergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotNaiveFloatCond(divergentOffset)
	}
}

func BenchmarkMandelbrotOptimizedEscape1_Divergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotOptimizedEscape1(divergentOffset)
	}
}

func BenchmarkMandelbrotOptimizedEscape2_Divergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotOptimizedEscape2(divergentOffset)
	}
}

func BenchmarkMandelbrotOptimizedEscape3_Divergent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mandelbrotOptimizedEscape3(divergentOffset)
	}
}
