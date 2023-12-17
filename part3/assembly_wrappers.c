#ifndef ASSEMBLY_FUNCTIONS_H
#define ASSEMBLY_FUNCTIONS_H

#include <stdint.h>

extern void ASM_MOVAllBytesASM(uint8_t* buffer, uint64_t count);
extern void ASM_NOPAllBytesASM(uint64_t count);
extern void ASM_CMPAllBytesASM(uint64_t count);
extern void ASM_DECAllBytesASM(uint64_t count);

#endif /* ASSEMBLY_FUNCTIONS_H */

// Wrapper for MOVAllBytesASM
void MOVAllBytesASM_CWrapper(uint8_t* buffer, uint64_t count) {
    register uint8_t* x1 asm("x1") = buffer;
    register uint64_t x2 asm("x2") = count;
    ASM_MOVAllBytesASM(x1, x2);
}

// Wrapper for NOPAllBytesASM
void NOPAllBytesASM_CWrapper(uint64_t count) {
    // Assembly function expects:
    // x0 = count
    register uint64_t x0 asm("x0") = count;
    ASM_NOPAllBytesASM(x0);
}

// Wrapper for CMPAllBytesASM
void CMPAllBytesASM_CWrapper(uint64_t count) {
    register uint64_t x0 asm("x0") = count;
    ASM_CMPAllBytesASM(x0);
}

// Wrapper for DECAllBytesASM
void DECAllBytesASM_CWrapper(uint64_t count) {
    register uint64_t x0 asm("x0") = count;
    ASM_DECAllBytesASM(x0);
}

//void main() {
//  // calls to assembly here
//}