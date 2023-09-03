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
	// Displacement values for 0b01 and 0b10 are handled directly
}

func disassemble(reader *bytes.Reader) string {
	var out bytes.Buffer
	out.WriteString("bits 16\n")

	b, err := reader.ReadByte()
	for err == nil {
		switch {
		// mov reg/mem to/from reg.
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

				// Memory to register, 8-bit displacement.
			case 0b01:
				disp8, _ := reader.ReadByte()
				destination = formatAddress(baseAddresses[secondByte&0b111], int16(disp8))

			// Memory to register, 16-bit displacement.
			case 0b10:
				lowDisp, _ := reader.ReadByte()
				highDisp, _ := reader.ReadByte()
				disp16 := int16(highDisp)<<8 | int16(lowDisp)
				destination = formatAddress(baseAddresses[secondByte&0b111], disp16)

			default:
				log.Fatalf("unsupported mod: %b", mod)
			}

			// Swap if d = 1 to handle `mov [memory], reg`.
			if d == 1 {
				destination, source = source, destination
			}

			out.WriteString("mov ")
			out.WriteString(destination)
			out.WriteString(", ")
			out.WriteString(source)
			out.WriteByte('\n')

		// mov immediate to reg.
		case b>>4 == 0b00001011:
			w := (b >> 3) & 1
			reg := b & 0b111

			secondByte, err := reader.ReadByte()
			if err != nil {
				log.Fatalf("error reading byte: %v", err)
			}

			if w == 0 {
				// Cast to int8 for signed conversion
				out.WriteString(fmt.Sprintf("mov %s, %d", regs[w][reg], int8(secondByte)))
				out.WriteByte('\n')
			} else {
				thirdByte, err := reader.ReadByte()
				if err != nil {
					log.Fatalf("error reading byte: %v", err)
				}
				// Convert two bytes to int16 for signed conversion
				value := int16(thirdByte)<<8 | int16(secondByte)
				out.WriteString(fmt.Sprintf("mov %s, %d", regs[w][reg], value))
				out.WriteByte('\n')
			}
		}

		b, err = reader.ReadByte()
	}

	return out.String()
}

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

func formatAddress(base string, disp int16) string {
	if disp == 0 {
		return "[" + base + "]"
	} else if disp > 0 {
		return fmt.Sprintf("[%s + %d]", base, disp)
	}
	return fmt.Sprintf("[%s - %d]", base, -disp)
}
