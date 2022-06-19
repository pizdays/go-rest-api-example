package util

import "math"

// RoundTo rounds v to n decimal places.
func RoundTo(v float64, n int) float64 {
	return math.Round(v*math.Pow10(n)) / math.Pow10(n)
}
