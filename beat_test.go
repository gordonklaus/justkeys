package main

import (
	"fmt"
	"testing"
)

func TestBeatSpectrum(t *testing.T) {
	A := make([]float64, 20)
	for i := range A {
		A[i] = 1 / float64(i+1)
	}

	// for a := 1; a < 6; a++ {
	// 	for b := a; b < 4*a; b++ {
	// 		if gcd(a, b) != 1 {
	// 			continue
	// 		}
	// 		fmt.Printf("%d:%d   ", a, b)
	// 		fmt.Printf("(%0.2f)\n", roughness(A, a, b))
	// 	}
	// }

	for a := 1; a < 11; a++ {
		for b := a; b < 2*a; b++ {
			for c := b; c < 2*a; c++ {
				if gcd(a, gcd(b, c)) != 1 {
					continue
				}
				fmt.Printf("%d:%d:%d   ", a, b, c)
				fmt.Printf("(%0.2f)\n", roughness(A, a, b, c))
			}
		}
	}
}

func TestGCD2(t *testing.T) {
	// x, y, gcd, a_gcd, b_gcd := gcd2(2, -3)
	// fmt.Println(x, y, gcd, a_gcd, b_gcd)
}

func TestLCM(t *testing.T) {
	if lcm(3, 4) != 12 {
		t.Fatal()
	}
}

func TestRatioN(t *testing.T) {
	N := ratioN([]ratio{{2, 3}, {4, 5}})
	if N[0] != 10 || N[1] != 12 {
		t.Fatal(N)
	}
}
