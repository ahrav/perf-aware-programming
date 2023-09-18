package main

import (
	"testing"
)

func BenchmarkReadGeoPairsFromFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = readGeoPairsFromFile("testdata/data.json")
	}
}

func BenchmarkCalcHaversineDistanceAvg(b *testing.B) {
	pairs, _ := readGeoPairsFromFile("testdata/data.json")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calcHaversineDistanceAvg(pairs)
	}
}
