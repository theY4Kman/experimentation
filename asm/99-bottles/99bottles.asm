; 99 Bottles, NASM-style
; by Zach "theY4Kman" Kanzler

segment .data
	bottles db " bottles of beer on the wall."

segment .text
	global _start
	
_start:
	mov	edi, 99		; start the loop at 99, of course!
	
	.loop:
	mov	eax, 4		; sys_write is syscall number four
	mov	ebx, 1		; stdout is output number one
	mov	ecx, about	; the string we wish to print
	mov	edx, [ablen]	; the length of that string
	int	80h		; perform the syscall