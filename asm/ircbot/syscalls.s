; ------------------------------------------------------
; socket_send: socket.h send() syscall
; params: (int fd, void *buff, size_t len, unsigned long flags)
; ------------------------------------------------------
socket_send:
    mov eax, 102    ; socketcall
    mov ebx, 9      ; send
    mov ecx, esp    ; unsigned long *args
    int 80h         ; call it in
    ret
    
; ------------------------------------------------------
; sys_brk: modifies the size of a program's heap
; ------------------------------------------------------
sys_brk:
    mov eax, 45     ; sys_brk
    int 80h         ; let her rip
    ret
    
sys_write:
    mov eax, 4      ; sys_write
    mov ebx, 1      ; stdout
    int 80h         ; perform the syscall
    ret
    
sys_exit:
    mov eax, 1      ; sys_exit
    mov ebx, 2      ; return code 0
    int 80h         ; end proggy
