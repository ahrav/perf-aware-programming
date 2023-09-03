package main

import (
	"bytes"
	"testing"
)

func TestDisassemble_SingleRegister(t *testing.T) {
	// Test data: Assume the first byte is MOV (for simplicity), followed by another byte.
	input := []byte{0b10001001, 0b11011001}

	expectedOutput := "bits 16\nmov CX, BX\n"

	reader := bytes.NewReader(input)
	result := disassemble(reader)

	if result != expectedOutput {
		t.Errorf("Expected '%s' but got '%s'", expectedOutput, result)
	}
}

func TestDisassemble_MultipleRegisters(t *testing.T) {
	input := []byte{
		0b10001001, 0b11011001,
		0b10001000, 0b11100101,
		0b10001001, 0b11011010,
		0b10001001, 0b11011110,
		0b10001001, 0b11111011,
		0b10001000, 0b11001000,
		0b10001000, 0b11101101,
		0b10001001, 0b11000011,
		0b10001001, 0b11110011,
		0b10001001, 0b11111100,
		0b10001001, 0b11000101,
	}

	expectedOutput := `bits 16
mov CX, BX
mov CH, AH
mov DX, BX
mov SI, BX
mov BX, DI
mov AL, CL
mov CH, CH
mov BX, AX
mov BX, SI
mov SP, DI
mov BP, AX
`
	reader := bytes.NewReader(input)
	result := disassemble(reader)

	if result != expectedOutput {
		t.Errorf("Expected:\n%s\nBut got:\n%s", expectedOutput, result)
	}
}

func BenchmarkDisassemble(b *testing.B) {
	input := []byte{0b10001001, 0b11011001}

	// Pre-allocate readers.
	readers := make([]*bytes.Reader, b.N)
	for i := range readers {
		readers[i] = bytes.NewReader(input)
	}

	b.ResetTimer() // Reset the timer after the setup.

	for i := 0; i < b.N; i++ {
		disassemble(readers[i])
	}
}
