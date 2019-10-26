package main

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/mathext"
)

func roughness(A []float64, N ...int) float64 {
	sort.Ints(N)

	d := N[0]
	for _, n := range N[1:] {
		d = gcd(d, n)
	}
	for i := range N {
		N[i] /= d
	}

	spec := beatSpectrum(A, N...)
	sum := 0.
	for _, a := range spec {
		sum += math.Sqrt(a)
	}
	return sum
}

func beatSpectrum(A []float64, N ...int) map[int]float64 {
	spectrum := make(map[int]float64, len(A)*len(N))
	for _, n := range N {
		for k, a := range A {
			spectrum[(k+1)*n] += a
		}
	}

	diffSpectrum := make(map[int]float64, len(spectrum)*(len(spectrum)-1)/2)

	for n1, a1 := range spectrum {
		for n2, a2 := range spectrum {
			if n1 < n2 {
				d := beatAmplitude(a1, a2) * dissonance(math.Log2(float64(n2)/float64(n1)))
				diffSpectrum[n2-n1] += d * d
			}
			// a := N[i]

			// e, f, gcd, a_gcd, b_gcd := gcd2(a, -b)
			// for c := gcd; c <= len(C); c += gcd {
			// 	// solutions x, y to ax-by=c
			// 	x, y := c*e/gcd, c*f/gcd
			// 	for ; x >= 0 && y >= 0; x, y = x-b_gcd, y-a_gcd {
			// 	}
			// 	for ; x <= len(A) && y <= len(A); x, y = x+b_gcd, y+a_gcd {
			// 		if x <= 0 || y <= 0 {
			// 			continue
			// 		}
			// 		C[c-1] = math.Max(C[c-1], math.Min(A[x-1], A[y-1]))
			// 	}

			// 	// solutions x, y to ax-by=-c
			// 	x, y = -c*e/gcd, -c*f/gcd
			// 	for ; x >= 0 && y >= 0; x, y = x-b_gcd, y-a_gcd {
			// 	}
			// 	for ; x <= len(A) && y <= len(A); x, y = x+b_gcd, y+a_gcd {
			// 		if x <= 0 || y <= 0 {
			// 			continue
			// 		}
			// 		C[c-1] = math.Max(C[c-1], math.Min(A[x-1], A[y-1]))
			// 	}
			// }
		}
	}
	return diffSpectrum
}

func dissonance(dp float64) float64 {
	x := 20 * dp
	return x * math.Exp(-x) * math.E
}

func beatAmplitude(a1, a2 float64) float64 {
	// return math.Min(a1, a2)
	// return a1 * a2 * math.Hypot(a1, a2)
	// return math.Hypot(a1, a2)

	// Avoid floating point rounding errors.
	if math.Abs(math.Log10(a1/a2)) > 10 {
		return math.Min(a1, a2)
	}

	meanSquare := a1*a1 + a2*a2

	m := 4 * a1 * a2 / ((a1 + a2) * (a1 + a2))
	squareMean := (a1 + a2) * mathext.CompleteE(m) / (math.Pi / 2)
	squareMean *= squareMean

	stddev := math.Sqrt(meanSquare - squareMean)
	return stddev
}

func gcd2(a, b int) (e, f, gcd, a_gcd, b_gcd int) {
	s, old_s := 0, 1
	t, old_t := 1, 0
	r, old_r := b, a

	for r != 0 {
		q := old_r / r
		old_r, r = r, old_r-q*r
		old_s, s = s, old_s-q*s
		old_t, t = t, old_t-q*t
	}

	e, f = old_s, old_t
	gcd = old_r
	if gcd < 0 {
		e, f, gcd = -e, -f, -gcd
	}
	a_gcd, b_gcd = fix_sign(t, a), fix_sign(s, b)
	return
}

func fix_sign(x, y int) int {
	if (x < 0) != (y < 0) {
		return -x
	}
	return x
}
