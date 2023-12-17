package main

import "testing"

// func BenchmarkMovAllBytes(b *testing.B) {
// 	buffer := make([]byte, 100) // Example buffer
// 	for i := 0; i < b.N; i++ {
// 		MovAllBytes(buffer, 100)
// 	}
// }

func BenchmarkNopAllBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NopAllBytes(100)
	}
}

func BenchmarkCmpAllBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CmpAllBytes(100)
	}
}

func BenchmarkDecAllBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecAllBytes(100)
	}
}
