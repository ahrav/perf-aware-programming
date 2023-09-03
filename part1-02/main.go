package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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

// regs maps the register index to the string representation of the register.
var regs = [][]string{
	{"al", "cl", "dl", "bl", "ah", "ch", "dh", "bh"},
	{"ax", "cx", "dx", "bx", "sp", "bp", "si", "di"},
}

var addressingModes = map[byte]map[byte]string{
	0b00: {
		0b000: "[bx + si]",
		0b001: "[bx + di]",
		0b010: "[bp + si]",
		0b011: "[bp + di]",
		0b100: "[si]",
		0b101: "[di]",
		0b110: "[bp]",
		0b111: "[bx]",
	},
	// Displacement values for 0b01 and 0b10 are handled directly.
}

func disassemble(reader *bytes.Reader) string {
	var out bytes.Buffer
	out.WriteString("bits 16\n")

	b, err := reader.ReadByte()
	for err == nil {
		switch {
		// Handle 'move reg/mem to/from reg' instructions.
		case b>>2 == 0b00100010:
			d := (b >> 1) & 1
			w := b & 1

			secondByte, err := reader.ReadByte()
			if err != nil {
				log.Fatalf("error reading byte: %v", err)
			}

			mod := (secondByte >> 6) & 0b11
			rm := secondByte & 0b111

			destination := ""
			source := regs[w][(secondByte>>3)&0b111]

			switch mod {
			// Register to register.
			case 0b11:
				destination = regs[w][rm]
				// Memory to register, no displacement.
			case 0b00:
				destination = addressingModes[mod][rm]
				// Memory to register, 8-bit displacement and 16-bit displacement.
			case 0b01, 0b10:
				destination = formatAddress(baseAddresses[secondByte&0b111], dispFromReader(reader, mod))
			default:
				out.WriteString("unsupported mod: ")
				out.WriteByte('0' + mod)
				out.WriteByte('\n')
			}

			// Swap if d == 1 to handle 'mov [memory], reg'.
			if d == 1 {
				destination, source = source, destination
			}

			out.WriteString("mov ")
			out.WriteString(destination)
			out.WriteString(", ")
			out.WriteString(source)
			out.WriteByte('\n')

			// Handle 'move immediate to reg/mem' instructions.
		case b>>4 == 0b00001011:
			w := (b >> 3) & 1
			reg := b & 0b111

			if w == 0 {
				// 8-bit immediate value.
				secondByte, _ := reader.ReadByte()
				out.WriteString("mov ")
				out.WriteString(regs[w][reg])
				out.WriteString(", ")
				// Convert to string representation using Itoa
				out.WriteString(strconv.Itoa(int(int8(secondByte))))
				out.WriteByte('\n')
			} else {
				// 16-bit immediate value.
				secondByte, _ := reader.ReadByte()
				thirdByte, _ := reader.ReadByte()
				value := int16(thirdByte)<<8 | int16(secondByte)
				out.WriteString("mov ")
				out.WriteString(regs[w][reg])
				out.WriteString(", ")
				// Convert to string representation using Itoa.
				out.WriteString(strconv.Itoa(int(value)))
				out.WriteByte('\n')
			}
		}

		b, err = reader.ReadByte()
	}

	return out.String()
}

// baseAddresses maps the base address to the string representation of the address.
var baseAddresses = map[byte]string{
	0b000: "bx + si",
	0b001: "bx + di",
	0b010: "bp + si",
	0b011: "bp + di",
	0b100: "si",
	0b101: "di",
	0b110: "bp",
	0b111: "bx",
}

// formatAddress formats the address based on the base and displacement.
func formatAddress(base string, disp int16) string {
	var builder strings.Builder
	builder.WriteByte('[')
	builder.WriteString(base)
	if disp > 0 {
		builder.WriteString(" + ")
		builder.WriteString(strconv.Itoa(int(disp)))
	} else if disp < 0 {
		builder.WriteString(" - ")
		builder.WriteString(strconv.Itoa(int(-disp)))
	}
	builder.WriteByte(']')
	return builder.String()
}

// dispFromReader reads the displacement value from the reader based on the addressing mode.
func dispFromReader(reader *bytes.Reader, mod byte) int16 {
	switch mod {
	case 0b01:
		disp8, _ := reader.ReadByte()
		return int16(disp8)
	case 0b10:
		lowDisp, _ := reader.ReadByte()
		highDisp, _ := reader.ReadByte()
		return int16(highDisp)<<8 | int16(lowDisp)
	default:
		return 0
	}
}
