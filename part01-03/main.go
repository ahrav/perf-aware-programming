package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/exp/constraints"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("dissasebling: ", os.Args[0])
	}

	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	res, err := disassembleFile(b)
	if err != nil {
		log.Fatalf("error disassembling file: %v", err)
	}
	fmt.Println(res)
}

type peekableByteReader struct {
	*bytes.Reader
}

func newPeekableBytReader(b []byte) *peekableByteReader {
	return &peekableByteReader{Reader: bytes.NewReader(b)}
}

// peek reads the next n bytes without advancing the position.
func (pr *peekableByteReader) peek(n int) ([]byte, error) {
	// Store the current position.
	pos, _ := pr.Seek(0, io.SeekCurrent)

	// Read the next n bytes.
	buf := make([]byte, n)
	if _, err := pr.Read(buf); err != nil && err != io.EOF {
		return nil, err
	}

	// Reset position.
	_, _ = pr.Seek(pos, io.SeekStart)

	return buf, nil
}

func (pr *peekableByteReader) readUint16W(wide bool) uint16 {
	if wide {
		return pr.readUint16()
	}

	readByte, _ := pr.ReadByte()
	return uint16(readByte)
}

func (pr *peekableByteReader) readInt16() int16 {
	result := pr.readUint16()
	return int16(result)
}

func (pr *peekableByteReader) readUint16() uint16 {
	b, err := pr.readN(2)
	if err != nil {
		log.Fatalf("error reading bytes: %v", err)
	}

	low := uint16(b[0])
	high := uint16(b[1])
	return (high << 8) | low
}

func (pr *peekableByteReader) readN(n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("non-positive n")
	}

	buf := make([]byte, n)
	read, err := pr.Reader.Read(buf)
	if read != n {
		return nil, fmt.Errorf("not enough bytes")
	}
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (pr *peekableByteReader) readUint8() uint8 {
	byteVal, err := pr.ReadByte()
	if err != nil {
		log.Fatalf("error reading byte: %v", err)
	}
	return byteVal
}

func (pr *peekableByteReader) readInt8() int8 {
	result, err := pr.ReadByte()
	if err != nil {
		log.Fatalf("error reading byte: %v", err)
	}
	return int8(result)
}

// instruction defines the contract for x86 instructions that can be disassembled
// into their human-readable representations. Implementers of this interface
// should provide a disassemble method that returns the assembly code equivalent
// of the binary-encoded instruction.
type instruction interface {
	disassemble() string
}

func disassembleFile(b []byte) (string, error) {
	var output bytes.Buffer
	output.WriteString("bits 16\n")

	reader := newPeekableBytReader(b)
	for reader.Len() > 0 {
		fmt.Printf("reader len: %d\n", reader.Len())
		b, err := reader.peek(min(4, reader.Len()))
		if err != nil {
			return "", err
		}

		firstByte := b[0]
		fmt.Println(firstByte)
		var instr instruction
		switch {
		// Handle move instruction.
		case isMovOp(firstByte):
			instr = decodeMov(reader)
			// Handle arithmetic instruction.
		case isArithmetic(firstByte):
			instr = decodeArithmetic(reader)
			// Handle jump/loop instruction.
		case isJumpOrLoopOp(firstByte):
			b, err := reader.ReadByte()
			if err != nil {
				return "", err
			}
			j := jumpOrLoop{op: b, inc: reader.readInt8()}
			instr = &j
		default:
			log.Fatalf("unsupported instruction opcode: %b", firstByte)
		}

		output.WriteString(instr.disassemble())
		output.WriteByte('\n')
	}

	return output.String(), nil
}

// isMovOp determines if the given byte represents a MOV operation.
func isMovOp(b byte) bool {
	return (b>>2) == 0b100010 || (b>>1) == 0b1100011 || (b>>4) == 0b1011 || (b>>1) == 0b1010000 || (b>>1) == 0b1010001 || b == 0b10001110 || b == 0b10001100
}

// isArithmetic determines if the given byte represents any arithmetic operation (ADD, SUB, CMP).
func isArithmetic(b byte) bool {
	return isAddOp(b) || isSubOp(b) || isCmpOp(b) || (b>>2) == 0b100000
}

// isAddOp determines if the given byte represents an ADD operation.
func isAddOp(b byte) bool {
	return (b>>2) == 0b0 || (b>>1) == 0b10
}

// isSubOp determines if the given byte represents a SUB operation.
func isSubOp(b byte) bool {
	return (b>>2) == 0b1010 || (b>>1) == 0b10110
}

// isCmpOp determines if the given byte represents a CMP operation.
func isCmpOp(b byte) bool {
	return (b>>2) == 0b1110 || (b>>1) == 0b11110
}

// isJumpOrLoopOp determines if the given byte represents a JUMP or LOOP operation.
func isJumpOrLoopOp(b byte) bool {
	return (b>>4) == 0b0111 || (b>>2) == 0b111000
}

// movType represents different types of MOV operations.
type movType uint8

const (
	movInvalid                     movType = iota // Invalid MOV type
	movRegisterMemToFromRegister                  // MOV from/to memory/register
	movImmediateToRegister                        // MOV immediate value to register
	movImmediateToMemoryOrRegister                // MOV immediate value to memory/register
	movMemoryToAccumulator                        // MOV from memory to accumulator
	movAccumulatorToMemory                        // MOV from accumulator to memory
)

// mov represents a MOV instruction with its type, data, and common opcode fields.
type mov struct {
	typ movType // Type of the MOV operation
	d   byte    // Direction bit
	w   byte    // Operand size bit

	common        // Common fields shared across opcodes
	data   uint16 // Data associated with the MOV operation (e.g., immediate value)
}

// common represents the shared fields present in many opcodes.
type common struct {
	mod  byte  // Mode field, typically used to specify addressing mode
	reg  byte  // Register field, often used to specify source/destination register
	rm   byte  // Register/memory field, often indicates a specific register or memory address
	disp int16 // Displacement value, typically an offset used in addressing
}

func (m *mov) disassemble() string {
	switch m.typ {
	case movRegisterMemToFromRegister:
		src := m.regName(m.w == 1)
		tgt := m.rmName(m.w == 1)

		if m.d == 1 {
			src, tgt = tgt, src
		}

		return fmt.Sprintf("mov %s, %s", tgt, src)
	case movImmediateToRegister:
		return fmt.Sprintf("mov %s, %d", m.regName(m.w == 1), m.data)
	case movImmediateToMemoryOrRegister:
		if m.mod == 0b11 {
			return fmt.Sprintf("mov %s, %d", m.rmName(m.w == 1), m.data)
		}

		dataT := "byte"
		if m.w == 1 {
			dataT = "word"
		}
		return fmt.Sprintf("mov %s, %s %d", m.rmName(m.w == 1), dataT, m.data)
	case movMemoryToAccumulator:
		return fmt.Sprintf("mov ax, [%d]", m.data)
	case movAccumulatorToMemory:
		return fmt.Sprintf("mov [%d], ax", m.data)
	default:
		log.Fatalf("invalid mov type: %d", m.typ)
		return ""
	}
}

// regs maps the register index to the string representation of the register.
var regs = map[bool][]string{
	false: {"al", "cl", "dl", "bl", "ah", "ch", "dh", "bh"},
	true:  {"ax", "cx", "dx", "bx", "sp", "bp", "si", "di"},
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

// determines the register name based on the register index and operand size.
func (c *common) regName(wide bool) string {
	return regs[wide][c.reg]
}

func (c *common) rmName(wide bool) string {
	switch c.mod {
	case 0b11:
		return regs[wide][c.rm]
	case 0b00:
		if c.rm == 0b110 {
			return fmt.Sprintf("[%d]", uint16(c.disp))
		}
		return fmt.Sprintf("[%s]", baseAddresses[c.rm])
	case 0b01, 0b10:
		base := baseAddresses[c.rm]
		dispStr := ""
		if c.disp < 0 {
			dispStr = fmt.Sprintf(" - %d", -c.disp)
		} else if c.disp > 0 {
			dispStr = fmt.Sprintf(" + %d", c.disp)
		}
		return fmt.Sprintf("[%s%s]", base, dispStr)
	default:
		panic("invalid mod value")
	}
}

// decodeMov decodes the MOV instruction from a stream of bytes provided by a peekableByteReader.
// It identifies the type of MOV operation based on the opcode and populates the mov struct accordingly.
func decodeMov(r *peekableByteReader) *mov {
	firstByte, err := r.ReadByte()
	if err != nil {
		log.Fatalf("error reading byte: %v", err)
	}

	instruction := mov{}

	switch {
	// Decode MOV from/to memory/register operation.
	case (firstByte >> 2) == 0b100010:
		instruction.typ = movRegisterMemToFromRegister
		instruction.d = (firstByte >> 1) & 1
		instruction.w = firstByte & 1
		instruction.common = decodeCommon(r)

	// Decode MOV immediate value to register.
	case (firstByte >> 4) == 0b1011:
		instruction.typ = movImmediateToRegister
		instruction.w = (firstByte >> 3) & 1
		instruction.common.reg = firstByte & 0b111
		instruction.data = r.readUint16W(instruction.w == 1)

	// Decode MOV immediate value to memory/register.
	case (firstByte >> 1) == 0b1100011:
		instruction.typ = movImmediateToMemoryOrRegister
		instruction.w = firstByte & 1
		instruction.common = decodeCommon(r)
		instruction.data = r.readUint16W(instruction.w == 1)

	// Decode MOV from memory to accumulator.
	case (firstByte >> 1) == 0b1010000:
		instruction.typ = movMemoryToAccumulator
		instruction.data = r.readUint16()

	// Decode MOV from accumulator to memory.
	case (firstByte >> 1) == 0b1010001:
		instruction.typ = movAccumulatorToMemory
		instruction.data = r.readUint16()

	default:
		log.Fatalf("unsupported instruction opcode: %b", firstByte)
	}

	return &instruction
}

// decodeCommon decodes common opcode fields shared across multiple instruction types from
// the provided byte stream. It decodes mode, register, and optional displacement values.
func decodeCommon(r *peekableByteReader) common {
	instruction := common{}
	b1, err := r.ReadByte()
	if err != nil {
		log.Fatalf("error reading byte: %v", err)
	}

	instruction.mod = (b1 >> 6) & 0b11
	instruction.reg = (b1 >> 3) & 0b111
	instruction.rm = b1 & 0b111

	// Decode displacement based on mod value.
	if instruction.mod == 0b01 {
		b, _ := r.ReadByte()
		instruction.disp = int16(int8(b))
	} else if instruction.mod == 0b10 || (instruction.mod == 0b00 && instruction.rm == 0b110) {
		instruction.disp = r.readInt16()
	}

	return instruction
}

// arithmeticType represents the type of arithmetic operation to be performed.
// It is used to categorize the different ways arithmetic can be applied within
// x86 assembly operations.
type arithmeticType uint8

const (
	// arithmeticInvalid is an undefined or unrecognized arithmetic type.
	arithmeticInvalid arithmeticType = iota

	// arithmeticRegOrMemWithRegToEither represents an operation between
	// a register/memory and a register, where the result can be stored
	// in either.
	arithmeticRegOrMemWithRegToEither

	// arithmeticImmediateToRegOrMem represents an operation where an
	// immediate value is used in conjunction with a register or memory value.
	arithmeticImmediateToRegOrMem

	// arithmeticImmediateToAccumulator represents an operation where
	// an immediate value is applied directly to the accumulator.
	arithmeticImmediateToAccumulator
)

// arithmeticOp defines specific arithmetic operations that can be executed.
type arithmeticOp uint8

const (
	// arithmeticInvalidOp is an undefined or unrecognized arithmetic operation.
	arithmeticInvalidOp arithmeticOp = iota

	// arithmeticAdd represents the addition operation.
	arithmeticAdd

	// arithmeticSub represents the subtraction operation.
	arithmeticSub

	// arithmeticCmp represents the comparison operation, essentially a
	// subtraction where the result is discarded but flags are set.
	arithmeticCmp
)

// arithmetic struct captures the details of an arithmetic instruction,
// including its type, specific operation, and other relevant metadata.
type arithmetic struct {
	typ arithmeticType // The type of arithmetic instruction.
	op  arithmeticOp   // The specific arithmetic operation (e.g., ADD, SUB).

	common           // Embedded type capturing common details across instructions.
	data      uint16 // Immediate data or offset used in the instruction.
	d         byte   // Direction flag: source and destination for operation.
	s         byte   // Sign flag: determines signedness of immediate data.
	w         byte   // Operand size: byte or word.
	firstByte byte   // The first byte of the instruction for reference.
}

// opName translates the byte representation of an arithmetic operation
// into its human-readable form. It determines the operation name based on
// the instruction's first byte and the `reg` field.
func (a *arithmetic) opName() string {
	b := a.firstByte
	switch {
	case isAddOp(b):
		return "add"
	case isSubOp(b):
		return "sub"
	case isCmpOp(b):
		return "cmp"
	default:
		switch a.reg {
		case 0b000:
			return "add"
		case 0b111:
			return "cmp"
		case 0b101:
			return "sub"
		default:
			log.Fatalf("invalid arithmetic operation: %b", b)
			return ""
		}
	}
}

func (a *arithmetic) disassemble() string {
	op := a.opName()

	switch a.typ {
	case arithmeticRegOrMemWithRegToEither:
		src := a.regName(a.w == 1)
		tgt := a.rmName(a.w == 1)

		if a.d == 1 {
			src, tgt = tgt, src
		}

		return fmt.Sprintf("%s %s, %s", op, tgt, src)
	case arithmeticImmediateToRegOrMem:
		if a.mod == 0b11 {
			return fmt.Sprintf("%s %s, %d", op, a.rmName(a.w == 1), a.data)
		}

		dataT := "byte"
		if a.w == 1 {
			dataT = "word"
		}

		return fmt.Sprintf("%s %s %s, %d", op, dataT, a.rmName(a.w == 1), a.data)
	case arithmeticImmediateToAccumulator:
		regName := "ax"
		if a.w == 0 {
			regName = "al"
		}
		return fmt.Sprintf("%s %s, %d", op, regName, a.data)
	default:
		log.Fatalf("invalid arithmetic type: %d", a.typ)
		return ""
	}
}

// decodeArithmetic decodes the arithmetic instruction from a stream of bytes provided by a peekableByteReader.
// It identifies the type of arithmetic operation based on the opcode and populates the arithmetic struct accordingly.
func decodeArithmetic(r *peekableByteReader) *arithmetic {
	firstByte, err := r.ReadByte()
	if err != nil {
		log.Fatalf("error reading byte: %v", err)
	}

	instruction := arithmetic{}
	instruction.firstByte = firstByte

	switch {
	case (firstByte >> 2) == 0b100000:
		instruction.typ = arithmeticImmediateToRegOrMem
		instruction.s = (firstByte >> 1) & 1
		instruction.w = (firstByte >> 0) & 1
		instruction.common = decodeCommon(r)
		instruction.data = r.readUint16W(instruction.s == 0 && instruction.w == 1)
	case (firstByte>>2)&1 == 0b1:
		instruction.typ = arithmeticImmediateToAccumulator
		instruction.w = (firstByte >> 0) & 1
		instruction.data = r.readUint16W(instruction.w == 1)
	case (firstByte>>2)&1 == 0b0:
		instruction.typ = arithmeticRegOrMemWithRegToEither
		instruction.d = (firstByte >> 1) & 1
		instruction.w = (firstByte >> 0) & 1
		instruction.common = decodeCommon(r)
	default:
		log.Fatalf("unsupported instruction opcode: %b", firstByte)
	}

	return &instruction
}

type jumpOrLoop struct {
	op  uint8
	inc int8
}

// byte begins with '0b0111'
var jumpLabels = []string{
	"jo",
	"jno",
	"jb",
	"jnb",
	"jz",
	"jnz",
	"jbe",
	"jnbe",
	"js",
	"jns",
	"jp",
	"jnp",
	"jl",
	"jnl",
	"jle",
	"jnle",
}

// byte begins with '0b111000'
var loopLabels = []string{
	"loopnz",
	"loopz",
	"loop",
	"jcxz",
}

func (j *jumpOrLoop) disassemble() string {
	return fmt.Sprintf("%s $+2%+d", j.opName(), j.inc)

}
func (j *jumpOrLoop) opName() string {
	if (j.op >> 4) == 0b0111 {
		return jumpLabels[j.op&0b1111]
	}

	return loopLabels[j.op&0b11]
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
