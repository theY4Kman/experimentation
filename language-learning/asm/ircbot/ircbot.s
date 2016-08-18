; Oh no, a simple IRC bot in ASM (NASM)
; By theY4Kman
%include "syscalls.s"
%include "string.s"

;extern getaddrinfo

segment .data
    about db "A simple IRC bot", 10, "Written by theY4Kman in Assembly", 13, 10, 0
    ablen dd $-about

;------------------------------------------------------------------------------
; IRC related stuff
;------------------------------------------------------------------------------
    IRC_USER db "USER yakbot yak asm :Yakbot Assembly IRC Bot", 13, 10, 0
    IRC_USER_LEN dd $-IRC_USER-1
    IRC_NICK db "NICK yakasmbot", 13, 10, 0 ; \r\n = 13, 10
    IRC_NICK_LEN dd $-IRC_NICK-1
    IRC_JOIN db "JOIN #iteam", 13, 10, 0
    IRC_JOIN_LEN dd $-IRC_JOIN-1
    
    IRC_PING db "PING :"

;------------------------------------------------------------------------------
; error messages
;------------------------------------------------------------------------------
    error_alloc db "Error allocating buffer", 10, 0
    error_alloc_len dd $-error_alloc
    error_recv db "*** Error receiving message from server", 10, 0
    error_recv_len dd $-error_recv

;------------------------------------------------------------------------------
; Miscellaneous
;------------------------------------------------------------------------------
    WE_GOT_PING db "Pinged! Too bad there ain't shit we can do about it", 10, 0

segment .bss
    buffer resb 512
    
segment .text
    global _start

_start:
    mov eax, 4          ; sys_write
    mov ebx, 1          ; stdout
    mov ecx, about
    mov edx, [ablen]
    int 80h             ; perform the syscall
    
    xor eax, eax        ; set registers to 0
    xor ebx, ebx
    
    push    byte 6      ; IPPROTO_TCP
    push    byte 1      ; SOCK_STREAM
    push    byte 2      ; PF_INET
    
    mov eax, 102        ; socketcall syscall
    inc ebx             ; set ebx to 1
    mov ecx, esp        ; setup arguments
    int 80h             ; perform the syscall
    
    add esp, 16         ; clean up the stack
    
    mov edi, eax        ; save the sockfd to EDI
    
    pop eax             ; program name
;    call getaddrinfo

; SYS_CONNECT {{{
    inc ebx             ; increment ebx to 2
    
; struct sockaddr {
    push    dword 0x0000    ; pad the address/port to 14 bytes
    push    dword 0x1007f   ; 66.150.219.5 (0x05DB9642) (irc.gamesurge.net) -- 0x34bfa5d8; irc.freenode.net
    push    word 0xb1a      ; port 6667
    push    word bx         ; AF_INET = 2
; }
    mov ecx, esp        ; save pointer to the struct above
    
; connect() args {
    push    dword 16    ; int addrlen
    push    ecx         ; struct sockaddr *uservaddr
    push    edi         ; int sockfd
; }
    
    mov eax, 102        ; socketcall
    inc ebx             ; set ebx to 3 (connect)
    mov ecx, esp        ; unsigned long *args
    
    int 80h             ; CALL THAT SUCKER IN
; }}}

;------------------------------------------------------------------------------
; allocate memory for buffer
;------------------------------------------------------------------------------
    xor ebx, ebx    ; reset ebx
    call    sys_brk
    
    cmp eax, -1
    jz  error_alloc
    
    add eax, 512    ; allocate 512 bytes
    mov ebx, eax
    call    sys_brk
    
    cmp eax, -1     ; sys_brk errors?
    jz  error_alloc
    
    cmp eax, buffer+1   ; see if buffer has grown
    jz  error_alloc
    
;------------------------------------------------------------------------------
; send preliminary messages to the IRC server
;------------------------------------------------------------------------------
; send() args {
    push    0           ; unsigned flags = 0
    push    dword [IRC_USER_LEN]; size_t len
    push    IRC_USER    ; void *buff
    push    edi         ; int fd
; }
    mov eax, 102    ; socketcall
    mov ebx, 9      ; send
    mov ecx, esp    ; unsigned long *args
    int 80h         ; call it in
    
; send() args {
    push    0           ; unsigned flags = 0
    push    dword [IRC_NICK_LEN]; size_t len
    push    IRC_NICK    ; void *buff
    push    edi         ; int fd
; }
    mov eax, 102    ; socketcall
    mov ebx, 9      ; send
    mov ecx, esp    ; unsigned long *args
    int 80h         ; call it in
    
; send() args {
    push    0           ; unsigned flags = 0
    push    dword [IRC_JOIN_LEN]; size_t len
    push    IRC_JOIN    ; void *buff
    push    edi         ; int fd
; }
    mov eax, 102    ; socketcall
    mov ebx, 9      ; send
    mov ecx, esp    ; unsigned long *args
    int 80h         ; call it in
    
;------------------------------------------------------------------------------
; main loop
;------------------------------------------------------------------------------
.while:
; recv() args {
    push    0       ; unsigned flags = 0
    push    512     ; size_t size = 512 (size of our buffer)
    push    buffer      ; void *ubuf
    push    edi     ; int fd
; }
    
    mov eax, 102    ; socketcall
    mov ebx, 10     ; recv
    mov ecx, esp    ; unsigned long *args
    int 80h     ; fucking int80. i keep forgetting to call it
    
    test    eax, eax
    jz  .error_recv ; if 0 (error occurred), exit the program
    
    mov esi, eax    ; save str_len result
    
    mov ecx, buffer ; print out buffer
    mov edx, esi
    call    sys_write
    
    mov eax, buffer
    mov ebx, IRC_PING
    call    strcmp      ; See if we're being PINGed
    
    test eax, eax       ; If so, PONG back!
    jnz .no_pong
    call pong
    
.no_pong:
    
    jmp .while
    
;------------------------------------------------------------------------------
; Executed when recv() returns 0
;------------------------------------------------------------------------------
.error_recv:
    mov ecx, error_recv
    mov edx, [error_recv_len]
    call    sys_write
    call    sys_exit
    
;------------------------------------------------------------------------------
; Executed when allocation of the buffer becomes erroneous
;------------------------------------------------------------------------------
.error_alloc:
    mov ecx, error_alloc
    mov edx, dword [error_alloc_len]
    call    sys_write

;------------------------------------------------------------------------------
; Executed when the bot receives a PING
;------------------------------------------------------------------------------
pong:
    mov ecx, WE_GOT_PING    ; print out buffer
    mov edx, esi
    call    sys_write
    ret
