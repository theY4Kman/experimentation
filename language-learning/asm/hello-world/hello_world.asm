; A simple Hello, World
; Uses  Paul Carter's ASM IO library
%include "ex/asm_io.inc"

segment .data
	hw_str db "Hello, World", 10, 0
	hw_len dd $-hw_str

segment .bss

segment .text
	global	_start
_start:
	mov	eax, 4 ; sys_write
	mov	ebx, 1 ; stdout file descriptor
	mov	ecx, hw_str
	mov	edx, [hw_len]
	
	int	80h ; kernel call
	
	mov	eax, 1 ; sys_exit
	mov	ebx, 0 ; return code 0
	int	80h ; again, call the kernel
