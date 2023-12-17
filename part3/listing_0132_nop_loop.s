.global _ASM_MOVAllBytesASM
.align 4
_ASM_MOVAllBytesASM:
    eor x0, x0, x0 // x0 = 0 (equivalent to xor rax, rax)
.loop1:
    strb w0, [x1, x0] // x1 is the destination (equivalent to rdx)
    add x0, x0, 1
    cmp x0, x2 // x2 is the count (equivalent to rcx)
    b.lo .loop1
    ret

.global _ASM_NOPAllBytesASM
.align 4
_ASM_NOPAllBytesASM:
    eor x0, x0, x0
.loop2:
    nop
    add x0, x0, 1
    cmp x0, x1
    b.lo .loop2
    ret

.global _ASM_CMPAllBytesASM
.align 4
_ASM_CMPAllBytesASM:
    eor x0, x0, x0
.loop3:
    add x0, x0, 1
    cmp x0, x1
    b.lo .loop3
    ret

.global _ASM_DECAllBytesASM
.align 4
_ASM_DECAllBytesASM:
.loop4:
    subs x0, x0, 1 // x0 is the counter (equivalent to rcx)
    b.ne .loop4
    ret
