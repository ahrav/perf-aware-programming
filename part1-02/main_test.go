package main

import (
	"bytes"
	"testing"
)

func TestDisassemble(t *testing.T) {
	// Input byte slice.
	input := []byte{
		0b10001001, 0b11011110,
		0b10001000, 0b11000110,
		0b10110001, 0b00001100,
		0b10110101, 0b11110100,
		0b10111001, 0b00001100, 0b00000000,
		0b10111001, 0b11110100, 0b11111111,
		0b10111010, 0b01101100, 0b00001111,
		0b10111010, 0b10010100, 0b11110000,
		0b10001010, 0b00000000,
		0b10001011, 0b00011011,
		0b10001011, 0b01010110,
		0b00000000, 0b10001010,
		0b01100000, 0b00000100,
		0b10001010, 0b10000000,
		0b10000111, 0b00010011,
		0b10001001, 0b00001001,
		0b10001000, 0b00001010,
		0b10001000, 0b01101110,
		0b00000000,
	}

	// Expected output string.
	expectedOutput := `bits 16
mov si, bx
mov dh, al
mov cl, 12
mov ch, -12
mov cx, 12
mov cx, -12
mov dx, 3948
mov dx, -3948
mov al, [bx + si]
mov bx, [bp + di]
mov dx, [bp]
mov ah, [bx + si + 4]
mov al, [bx + si + 4999]
mov [bx + di], cx
mov [bp + si], cl
mov [bp], ch
`

	reader := bytes.NewReader(input)
	result := disassemble(reader)

	if result != expectedOutput {
		t.Errorf("Expected:\n'%s'\nGot:\n'%s'", expectedOutput, result)
	}
}

func BenchmarkDisassemble(b *testing.B) {
	input := []byte{
		0b10001001, 0b11011110,
		0b10001000, 0b11000110,
		0b10110001, 0b00001100,
		0b10110101, 0b11110100,
		0b10111001, 0b00001100, 0b00000000,
		0b10111001, 0b11110100, 0b11111111,
		0b10111010, 0b01101100, 0b00001111,
		0b10111010, 0b10010100, 0b11110000,
		0b10001010, 0b00000000,
		0b10001011, 0b00011011,
		0b10001011, 0b01010110,
		0b00000000, 0b10001010,
		0b01100000, 0b00000100,
		0b10001010, 0b10000000,
		0b10000111, 0b00010011,
		0b10001001, 0b00001001,
		0b10001000, 0b00001010,
		0b10001000, 0b01101110,
		0b00000000,
	}

	reader := bytes.NewReader(input)

	for i := 0; i < b.N; i++ {
		disassemble(reader)
		reader.Seek(0, 0)
	}
}
