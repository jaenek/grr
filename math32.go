package main

import "math"

func cos(f float32) float32 {
	return float32(math.Cos(float64(f)))
}

func sin(f float32) float32 {
	return float32(math.Sin(float64(f)))
}
