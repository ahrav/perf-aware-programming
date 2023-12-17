package main

/*
#cgo LDFLAGS: -L. -lassembly
#include <stdint.h>
extern void MOVAllBytesASM_CWrapper(uint8_t* buffer, uint64_t count);
extern void NOPAllBytesASM_CWrapper(uint64_t count);
extern void CMPAllBytesASM_CWrapper(uint64_t count);
extern void DECAllBytesASM_CWrapper(uint64_t count);
*/
import "C"
import "unsafe"

func MovAllBytes(buffer []byte, count uint64) {
	cBuffer := (*C.uint8_t)(unsafe.Pointer(&buffer[0])) // Get address of Go slice
	C.MOVAllBytesASM_CWrapper(cBuffer, C.uint64_t(count))
}

// TODO: This is not working correctly.
func NopAllBytes(count uint64) {
	C.NOPAllBytesASM_CWrapper(C.uint64_t(count))
}

// TODO: This is not working correctly.
func CmpAllBytes(count uint64) {
	C.CMPAllBytesASM_CWrapper(C.uint64_t(count))
}

func DecAllBytes(count uint64) {
	C.DECAllBytesASM_CWrapper(C.uint64_t(count))
}

func main() {
	buffer := make([]byte, 100) // Example buffer
	MovAllBytes(buffer, 100)
	NopAllBytes(100)
	CmpAllBytes(100)
	DecAllBytes(100)
}
