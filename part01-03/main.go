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
		// Handle add reg/mem to/from reg.
		case b>>2 == 0b00000000:
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
			case 0b11: // Register to register.
				destination = regs[w][rm]

			case 0b00:
				if rm == 0b110 { // Special case for direct address
					disp16 := mustReadInt16(reader)
					destination = "[" + strconv.Itoa(int(disp16)) + "]"
				} else {
					destination = addressingModes[mod][rm]
				}

			case 0b01: // Memory to register, 8-bit displacement.
				disp8, _ := reader.ReadByte()
				destination = formatAddress(baseAddresses[secondByte&0b111], int16(disp8))

			case 0b10: // Memory to register, 16-bit displacement.
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

			if mod == 0b11 {
				data := mustReadUint16W(w == 1, reader)
				out.WriteString("add ")
				out.WriteString(destination)
				out.WriteString(", ")
				out.WriteString(strconv.Itoa(int(data)))
				out.WriteByte('\n')
			} else {
				out.WriteString("add ")
				out.WriteString(destination)
				out.WriteString(", ")
				out.WriteString(source)
				out.WriteByte('\n')
			}

		// Handle add immediate to reg or memory.
		case b>>2 == 0b100000:
			w := b & 1
			s := (b >> 1) & 1

			b1, _ := reader.ReadByte()
			mod := (b1 >> 6) & 0b11
			regOpcode := (b1 >> 3) & 0b111 // Extract the 3 bits after mod
			rm := b1 & 0b111

			var destination string
			if mod == 0b00 && rm == 0b110 {
				disp16 := mustReadInt16(reader)
				destination = "[" + strconv.Itoa(int(disp16)) + "]"
			} else if mod == 0b01 {
				disp8 := mustReadInt8(reader)
				destination = formatAddress(baseAddresses[rm], int16(disp8))
			} else if mod == 0b10 {
				disp16 := mustReadInt16(reader)
				destination = formatAddress(baseAddresses[rm], disp16)
			} else if mod == 0b11 {
				destination = regs[w][rm]
			} else {
				destination = addressingModes[mod][rm]
			}

			data := mustReadUint16W(s == 0 && w == 1, reader)

			var operation string
			if regOpcode == 0b000 {
				operation = "add"
			} else if regOpcode == 0b101 {
				operation = "sub"
			} else {
				continue
			}

			out.WriteString(operation + " ")
			if mod != 0b11 && w == 1 {
				out.WriteString("word ")
			}
			out.WriteString(destination)
			out.WriteString(", ")
			out.WriteString(strconv.Itoa(int(data)))
			out.WriteByte('\n')

		// Handle add immediate to accumulator.
		case b>>2 == 0b000001:
			w := b & 1
			var data uint16

			if w == 1 { // word
				data = mustReadUint16(reader)
			} else { // byte
				data = uint16(mustReadInt8(reader))
			}

			// Accumulator register is either 'al' or 'ax' based on w.
			accumulator := regs[1][0] // default to 'ax'
			if w == 0 {
				accumulator = regs[0][0] // change to 'al'
			}

			out.WriteString("add ")
			out.WriteString(accumulator)
			out.WriteString(", ")
			out.WriteString(strconv.Itoa(int(data)))
			out.WriteByte('\n')

		// Handle sub reg/mem to/from reg.
		case b>>2 == 0b001010:
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
			case 0b11: // Register to register.
				destination = regs[w][rm]

			case 0b00:
				if rm == 0b110 { // Special case for direct address
					disp16 := mustReadInt16(reader)
					destination = "[" + strconv.Itoa(int(disp16)) + "]"
				} else {
					destination = addressingModes[mod][rm]
				}

			case 0b01: // Memory to register, 8-bit displacement.
				disp8, _ := reader.ReadByte()
				destination = formatAddress(baseAddresses[secondByte&0b111], int16(disp8))

			case 0b10: // Memory to register, 16-bit displacement.
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

			if mod == 0b11 {
				data := mustReadUint16W(w == 1, reader)
				out.WriteString("sub ")
				out.WriteString(destination)
				out.WriteString(", ")
				out.WriteString(strconv.Itoa(int(data)))
				out.WriteByte('\n')
			} else {
				out.WriteString("sub ")
				out.WriteString(destination)
				out.WriteString(", ")
				out.WriteString(source)
				out.WriteByte('\n')
			}

			// Handle sub immediate to accumulator.
		case b>>1 == 0b0010110:
			w := b & 1
			var data uint16

			if w == 1 { // word
				data = mustReadUint16(reader)
			} else { // byte
				data = uint16(mustReadInt8(reader))
			}

			// Accumulator register is either 'al' or 'ax' based on w.
			accumulator := regs[1][0] // default to 'ax'
			if w == 0 {
				accumulator = regs[0][0] // change to 'al'
			}

			out.WriteString("sub ")
			out.WriteString(accumulator)
			out.WriteString(", ")
			out.WriteString(strconv.Itoa(int(data)))
			out.WriteByte('\n')

		}

		b, err = reader.ReadByte()
	}

	return out.String()
}

func mustReadUint16W(wide bool, r *bytes.Reader) uint16 {
	if wide {
		return mustReadUint16(r)
	}

	readByte, _ := r.ReadByte()
	return uint16(readByte)
}

func mustReadInt16(r *bytes.Reader) int16 {
	result := mustReadUint16(r)
	return int16(result)
}

func mustReadUint16(r *bytes.Reader) uint16 {
	b := make([]byte, 2)
	_, err := r.Read(b)
	if err != nil {
		panic(err)
	}
	low := uint16(b[0])
	high := uint16(b[1])
	return (high << 8) | low
}

func mustReadInt8(r *bytes.Reader) int8 {
	result, _ := r.ReadByte()
	return int8(result)
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
		builder.WriteString(strconv.Itoa(int(disp))) // Adds the negative sign
	}
	builder.WriteByte(']')
	return builder.String()
}
