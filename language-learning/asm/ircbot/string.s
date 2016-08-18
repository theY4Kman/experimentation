; ------------------------------------------------------
; strlen: finds the length of a C string
; params: eax = char*
; return(eax): the string's length
; ------------------------------------------------------
strlen:
    mov ecx, eax    ; set ecx = eax
    
    .strlen_test:
    mov bl, [eax]
    test    bl, bl      ; sets the zero flag if (bl & bl) is 0
    jz  .strlen_end
    inc eax
    jmp .strlen_test
    
    .strlen_end:
    sub eax, ecx
    ret
    
; ------------------------------------------------------
; strcmp: compare two C strings
; params: eax: string 1
;         ebx: string 2
; return(eax): 0 if matching or non-zero if not.
;              when non-zero, returns the difference
;              between the two differing characters
; ------------------------------------------------------
strcmp:
    mov edx, eax    ; we'll use edx to store the initial value of eax
    xor ecx, ecx    ; and ecx to store the two characters
    
    .next:
    mov cl, [eax]   ; cl = char from string 1
    mov ch, [ebx]   ; ch = char from string 2
    cmp cl, ch
    jne .not_equal  ; if not equal, ABORT ABORT ABORT!
    
    test    cl, cl
    jz  .equal      ; when NULL is reached, they're equal!
    
    inc eax
    inc ebx
    
    jmp .next
    
    .equal:
    xor eax, eax    ; return 0
    ret
    
    .not_equal:
    mov eax, [eax]
    sub eax, ebx
    ret
    
; ------------------------------------------------------
; strncmp: compare two C strings
; params: eax: string 1
;         ebx: string 2
;         ecx: max amount of chars to check
; return(eax): a pointer to the differing character of
;              string 1
; ------------------------------------------------------
strncmp:
    mov edx, ecx    ; we'll use edx to store the count
    xor ecx, ecx    ; and ecx to store the two characters
    
    .next:
    test edx, edx
    jz .equal       ; The count has been met -- they are equal
    
    mov cl, [eax]   ; cl = char from string 1
    mov ch, [ebx]   ; ch = char from string 2
    cmp cl, ch
    jne .not_equal  ; if not equal, ABORT ABORT ABORT!
    
    test    cl, cl
    jz  .equal      ; when NULL is reached, they're equal!
    
    inc eax
    inc ebx
    inc edx
    
    jmp .next
    
    .equal:
    xor eax, eax    ; return 0
    ret
    
    .not_equal:
    ret
