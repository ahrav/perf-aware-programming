package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("dissasebling: ", os.Args[0])
	}

	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	reader := bytes.NewReader(b)
	fmt.Println(disassemble(reader))
}

var regs = [][]string{
	{"AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH"},
	{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI"},
}

func disassemble(reader *bytes.Reader) string {
	var out bytes.Buffer
	out.WriteString("bits 16\n")

	b, err := reader.ReadByte()
	for err == nil {
		// mov
		if b>>2 == 0b00100010 {
			d := (b >> 1) & 1
			w := b & 1

			secondByte, err := reader.ReadByte()
			if err != nil {
				log.Fatalf("error reading byte: %v", err)
			}

			if mod := (secondByte >> 6) & 0b11; mod != 0b11 {
				log.Fatalf("only reg to reg move supported, got mod: %b", mod)
			}

			src, dst := (secondByte>>3)&0b111, secondByte&0b111
			if d == 1 {
				src, dst = dst, src
			}

			out.WriteString("mov ")
			out.WriteString(regs[w][dst])
			out.WriteString(", ")
			out.WriteString(regs[w][src])
			out.WriteByte('\n')
		}

		b, err = reader.ReadByte()
	}

	return out.String()
}
