; Oh no, a simple IRC bot in ASM
; By theY4Kman
	;extern printf

segment .data
	output db "reg: %x", 0
	about db "A simple IRC bot", 10, "Written by theY4Kman in Assembly", 10, 0
	ablen dd $-about
	IRC_USER db "USER yakbot yak asm :Yakbot Assembly IRC Bot", 13, 10
	IRC_USER_LEN dd $-IRC_USER
	IRC_NICK db "NICK yakbot", 13, 10 ; \r\n = 13, 10
	IRC_NICK_LEN dd $-IRC_NICK
	IRC_JOIN db "JOIN #iteam", 13, 10
	IRC_JOIN_LEN dd $-IRC_JOIN

segment .bss
segment .text
	global _start

_start:
	mov	eax, 4		; sys_writerun
	mov	ebx, 1		; stdout
	mov	ecx, about
	mov	edx, [ablen]
	int	80h		; perform the syscall
	
	xor	eax, eax	; set registers to 0
	xor	ebx, ebx
	
	push	byte 6		; IPPROTO_TCP
	push	byte 1		; SOCK_STREAM
	push	byte 2		; PF_INET
	
	mov	eax, 102	; socketcall syscall
	inc	ebx		; set ebx to 1
	mov	ecx, esp	; setup arguments
	int	80h		; perform the syscall
	
	add	esp, 16		; clean up the stack
	
	mov	edi, eax	; save the sockfd to EDI
	
	;push	dword [eax]
	;push	output
	;call	printf

; SYS_CONNECT {{{
	inc	ebx		; increment ebx to 2
	
; struct sockaddr {
	push	dword 0		; pad the address/port to 14 bytes
	push	dword 0		; more padding
	push	dword 0x0100007f; 127.0.0.1
	push	word 0xa31f	; port 8099
	push	word bx		; AF_INET = 2
; }
	mov	ecx, esp	; save pointer to the struct above
	
; connect() args {
	push	dword 16	; int addrlen
	push	ecx		; struct sockaddr *uservaddr
	push	edi		; int sockfd
; }
	
	mov	eax, 102	; socketcall
	inc	ebx		; set ebx to 3 (connect)
	mov	ecx, esp	; unsigned long *args
	
	int	80h		; CALL THAT SUCKER IN
; }}}
	
	add	esp, 14		; clean up the stack: sockaddr struct
	
;	push	512 		; allocate 512 bytes as a buffer for recv()
;	call	malloc
;	mov	esi, eax	; save the pointer to our allocated memory
;	add	esp, 4		; clean up stack
	
; send() args {
	push	0		; unsigned flags = 0
	push	dword [IRC_USER_LEN]; size_t len
	push	IRC_USER	; void *buff
	push	edi		; int fd
; }
	mov	eax, 102	; socketcall
	mov	ebx, 9		; send
	mov	ecx, esp	; unsigned long *args
	int	80h		; call it in :|
	
; send() args {
	push	0		; unsigned flags = 0
	push	dword [IRC_NICK_LEN]; size_t len
	push	IRC_NICK	; void *buff
	push	edi		; int fd
; }
	mov	eax, 102	; socketcall
	mov	ebx, 9		; send
	mov	ecx, esp	; unsigned long *args
	int	80h		; call it in :|
	
; send() args {
	push	0		; unsigned flags = 0
	push	dword [IRC_JOIN_LEN]; size_t len
	push	IRC_JOIN	; void *buff
	push	edi		; int fd
; }
	mov	eax, 102	; socketcall
	mov	ebx, 9		; send
	mov	ecx, esp	; unsigned long *args
	int	80h		; call it in :|
	
	mov	esi, esp	; buffer on the stack. yuck.
	sub	esi, 512+16	; 512b + the 16b we push on later
	
; recv() args {
.while:	push	0		; unsigned flags = 0
	push	512		; size_t size = 512 (size of our buffer)
	push	esi		; void *ubuf
	push	edi		; int fd
; }
	
	mov	eax, 102	; socketcall
	mov	ebx, 10		; recv
	mov	ecx, esp	; unsigned long *args
	add	esp, 16		; clean up stack
	
	jmp	.while
	
	mov	eax, 1		; sys_exit
	mov	ebx, 0		; return code 0
	int	80h		; end proggy
	
send:
	mov	eax, 102	; socketcall
	mov	ebx, 9		; send
	mov	ecx, esp	; unsigned long *args
	int	80h		; call it in :|
	ret