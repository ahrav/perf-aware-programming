package prologue

import "testing"

func BenchmarkSingleScalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SingleScalar(4096, input)
	}
}

func BenchmarkUnroll2Scalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unroll2Scalar(4096, input)
	}
}

func BenchmarkDualScalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DualScalar(4096, input)
	}
}

func BenchmarkQuadScalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuadScalar(4096, input)
	}
}

func BenchmarkOctoScalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OctoScalar(4096, input)
	}
}

func BenchmarkQuadScalarPtr(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuadScalarPtr(4096, &input)
	}
}

func BenchmarkQuadScalarNoBoundsCheck(b *testing.B) {
	input := make([]uint32, 4096)
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuadScalarNoBoundsCheck(4096, input)
	}
}

func BenchmarkParallelDualScalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParallelDualScalar(4096, input)
	}
}

func BenchmarkParallelQuadScalar(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParallelQuadScalar(4096, input)
	}
}

func BenchmarkParallelSingleScalarOptimized(b *testing.B) {
	var input [4096]uint32
	for i := range input {
		input[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParallelSingleScalarOptimized(4096, input)
	}
}
