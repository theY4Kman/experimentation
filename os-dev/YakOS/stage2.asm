;******************************************************;
;            YakOS bootloader, stage 2
;               Written by theY4Kman
;******************************************************;
[ORG    0x0]
[BITS   16]

jmp     main

;******************************************************;
;           Methods
;******************************************************;
%include "useful_methods.asm"

;******************************************************;
;           Stage 2 loader entry point
;******************************************************;

main:
    cli
    push    cs
    pop     ds

    mov     si, msg
    call    Print

    cli
    hlt

;******************************************************;
;           Data Section
;******************************************************;
msg     db  "Preparing to load the operating system...", 13, 10, 0