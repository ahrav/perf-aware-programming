package prologue

import (
	"sync"
)

func SingleScalar(count uint32, input [4096]uint32) uint32 {
	var sum uint32 = 0
	for index := uint32(0); index < count; index++ {
		sum += input[index]
	}
	return sum
}

func Unroll2Scalar(count uint32, input [4096]uint32) uint32 {
	var sum uint32 = 0
	for index := uint32(0); index < count; index += 2 {
		sum += input[index]
		sum += input[index+1]
	}
	return sum
}

func DualScalar(count uint32, input [4096]uint32) uint32 {
	var sum1 uint32 = 0
	var sum2 uint32 = 0
	for index := uint32(0); index < count; index += 2 {
		sum1 += input[index]
		sum2 += input[index+1]
	}
	return sum1 + sum2
}

// Fastest so far.
func QuadScalar(count uint32, input [4096]uint32) uint32 {
	var sum1 uint32 = 0
	var sum2 uint32 = 0
	var sum3 uint32 = 0
	var sum4 uint32 = 0
	for index := uint32(0); index < count; index += 4 {
		sum1 += input[index]
		sum2 += input[index+1]
		sum3 += input[index+2]
		sum4 += input[index+3]
	}
	return sum1 + sum2 + sum3 + sum4
}

func QuadScalarNoBoundsCheck(count uint32, input []uint32) uint32 {
	var sum1, sum2, sum3, sum4 uint32
	length := len(input)
	for index := 0; index+3 < length && index < int(count); index += 4 {
		sum1 += input[index]
		sum2 += input[index+1]
		sum3 += input[index+2]
		sum4 += input[index+3]
	}
	return sum1 + sum2 + sum3 + sum4
}

func QuadScalarPtr(count uint32, input *[4096]uint32) uint32 {
	var sumA, sumB, sumC, sumD uint32
	count /= 4
	for i := uint32(0); i < count; i++ {
		sumA += input[4*i]
		sumB += input[4*i+1]
		sumC += input[4*i+2]
		sumD += input[4*i+3]
	}
	return sumA + sumB + sumC + sumD
}

// No bueno.
func OctoScalar(count uint32, input [4096]uint32) uint32 {
	var sum1 uint32 = 0
	var sum2 uint32 = 0
	var sum3 uint32 = 0
	var sum4 uint32 = 0
	var sum5 uint32 = 0
	var sum6 uint32 = 0
	var sum7 uint32 = 0
	var sum8 uint32 = 0
	for index := uint32(0); index < count; index += 8 {
		sum1 += input[index]
		sum2 += input[index+1]
		sum3 += input[index+2]
		sum4 += input[index+3]
		sum5 += input[index+4]
		sum6 += input[index+5]
		sum7 += input[index+6]
		sum8 += input[index+7]
	}
	return sum1 + sum2 + sum3 + sum4 + sum5 + sum6 + sum7 + sum8
}

func ParallelDualScalar(count uint32, input [4096]uint32) uint32 {
	mid := count / 2

	results := make(chan uint32, 2)

	go PartialSum(0, mid, input, results)
	go PartialSum(mid, count, input, results)

	// Collect results
	sum1 := <-results
	sum2 := <-results

	return sum1 + sum2
}

func PartialSum(start, end uint32, input [4096]uint32, results chan<- uint32) {
	var sum uint32 = 0
	for index := start; index < end; index++ {
		sum += input[index]
	}
	results <- sum
}

func ParallelQuadScalar(count uint32, input [4096]uint32) uint32 {
	quarter := count / 4

	results := make(chan uint32, 4)

	go PartialSum(0, quarter, input, results)
	go PartialSum(quarter, 2*quarter, input, results)
	go PartialSum(2*quarter, 3*quarter, input, results)
	go PartialSum(3*quarter, count, input, results)

	// Collect results
	sum1 := <-results
	sum2 := <-results
	sum3 := <-results
	sum4 := <-results

	return sum1 + sum2 + sum3 + sum4
}

func ParallelSingleScalarOptimized(count uint32, input [4096]uint32) uint32 {
	mid := count / 2

	var sum1, sum2 uint32

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for index := uint32(0); index < mid; index++ {
			sum1 += input[index]
		}
	}()

	go func() {
		defer wg.Done()
		for index := mid; index < count; index++ {
			sum2 += input[index]
		}
	}()

	wg.Wait()

	return sum1 + sum2
}
