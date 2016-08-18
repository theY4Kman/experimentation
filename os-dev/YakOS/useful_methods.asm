Print:
    lodsb  ;al, byte ptr ds:[si]
    test    al, al
    jz      .print_done
    mov     ah, 0x0e
    int     10h
    jmp     Print
.print_done:
    ret