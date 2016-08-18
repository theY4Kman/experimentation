;******************************************************;
;                 YakOS bootloader
; Written by theY4Kman, with much help from:
; http://www.brokenthorn.com/Resources/OSDevIndex.html
;******************************************************;

bits    16                          ; We are still in 16 bit Real Mode

org     0x0                         ; We are loaded by BIOS at 0x7C00

SECTION loader vstart=0x7C00
    jmp loader               ; jump over OEM block

;******************************************************;
;           OEM Parameter block
;******************************************************;

bpbOEM:                 DB "YakOS   "

bpbBytesPerSector:      DW 512
bpbSectorsPerCluster:   DB 1
bpbReservedSectors:     DW 1
bpbNumberOfFATs:        DB 2
bpbRootEntries:         DW 224
bpbTotalSectors:        DW 2880
bpbMedia:               DB 0xF0
bpbSectorsPerFAT:       DW 9
bpbSectorsPerTrack:     DW 18
bpbHeadsPerCylinder:    DW 2
bpbHiddenSectors:       DD 0
bpbTotalSectorsBig:     DD 0
bsDriveNumber:          DB 0
bsUnused:               DB 0
bsExtBootSignature:     DB 0x29
bsSerialNumber:         DD 0xa0a1a2a3
bsVolumeLabel:          DB "MOS FLOPPY "
bsFileSystem:           DB "FAT12   "

welcome                 DB "Congratulations on booting up YakOS! ", 13, 10, 0

;******************************************************;
;           Stage 1 Methods
;******************************************************;
print:
    lodsb
    test    al, al
    jz      .print_done
    mov     ah, 0x0e
    int     10h
    jmp     print
.print_done:
    ret

;******************************************************;
;           Bootloader Entry Point
;******************************************************;

loader:
    xor     ax, ax          ; null all our segments
    mov     ds, ax
    mov     es, ax

; setup our stack
    cli
    mov ax, 0x1000
    mov ss, ax
    mov sp, 0xFFFE
    sti

    mov     si, welcome
    call    print

.reset:
    mov     ah, 0           ; reset floppy disk function
    mov     dl, 0           ; drive 0 is floppy drive
    int     0x13            ; call BIOS
    jc      .reset          ; If Carry Flag (CF) is set, there was an error. Try resetting again

    mov     ax, 0x1000      ; we are going to read sector to into address 0x1000:0
    mov     es, ax
    xor     bx, bx

    mov     ah, 0x02        ; function 2, read from device
    mov     al, 1           ; read 1 sector
    mov     ch, 0           ; we are reading the second sector past us, so its still on track 1 (index 0)
    mov     cl, 2           ; sector to read (The second sector, index 2)
    mov     dh, 0           ; head number
    mov     dl, 0           ; drive number. Remember Drive 0 is floppy drive.
    int     13h             ; call BIOS - Read the sector

    jmp     0x1000:0x0      ; jump to execute the sector!

times 510 - ($-$$) db 0     ; We have to be 512 bytes. Clear the rest of the bytes with 0

dw 0xAA55                   ; Boot Signature


;******************************************************;
;           END OF SECTOR 1, BEGIN SECTOR 2
;******************************************************;
SECTION loader2 vstart=0x10000
jmp     loader2              ; jump to stage2, offsetted from 0x1000.
                             ; $ == location of current instruction
;******************************************************;
;           Stage 1, part 2 Methods
;******************************************************;
Print:
    lodsb  ;al, byte ptr ds:[si]
    test    al, al
    jz      .print_done
    mov     ah, 0x0e
    int     10h
    jmp     Print
.print_done:
    ret

;;; :TODO: FIX THIS METHOD!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
; liek nao kthx
ReadSectors:
    mov     es, bx          ; mov the location to store the data into es

    mov     ah, 0x02        ; function 2, read from device
    mov     dh, 0           ; head number
    mov     dl, 0           ; drive number. Remember Drive 0 is floppy drive.
    int     13h             ; call BIOS - Read the sector

    ret

secstage:   db  "Loading YakOS, second stage...", 13, 10, 0
fat12load:  db  "Retrieving YAKLDR.SYS...", 13, 10, 0
stage2name: db  "YAKLDR  SYS "

loader2:
    mov     ax, 0x1000
    mov     ds, ax                          ; Make DS 0x1000, so that all memory reads will use that offset

    mov     si, secstage
    call    Print

    mov     si, fat12load
    call    Print

    xor     cx, cx                          ; cl and, more importantly, ch == 0 (ch will be used to signify sector 0 later)
    xor     dx, dx

    mov     ax, 20h                         ; 32-byte (0x20) directory entry

    xor     ax, ax                          ; set ds to 0x0, so that we can use the OEM tags
    mov     ds, ax

    mul     WORD [bpbRootEntries]           ; Number of root entries
    div     WORD [bpbBytesPerSector]        ; Number of sectors used by root directory
    mov     cx, ax                          ; Store this in CX (we really only care about cl)

    mov     al, [bpbNumberOfFATs]           ; Number of FATs (usually 2)
    mul     WORD [bpbSectorsPerFAT]         ; Multiply that by the number of sectors per FAT == size of FATs in sectors
    add     ax, [bpbReservedSectors]        ; Add those reserved sectors, and we know where to start finding files (root directory)

    mov     ax, 0x1000
    mov     ds, ax                          ; set DS back

    mov     bx, 0x300                       ; Load root directory to 0x1000:0x0300
    call    ReadSectors

    xor     ax, ax                          ; set ds to 0x0, so that we can use the OEM tags
    mov     ds, ax

    mov     cx, [bpbRootEntries]
    mov     di, 0x300

    mov     ax, 0x1000
    mov     ds, ax                          ; set DS back
    mov     es, ax                          ; as well as ES, which will be used in the FAT12 search for YAKLDR.SYS

.srchloop:
    push    cx
    mov     cx, 11
    mov     si, stage2name
    push    di

rep cmpsb

    pop     di
    je      LOAD_FAT

    pop     cx
    add     di, 32
    loop    .srchloop

    jmp     SRCHFAILURE                     ; no more entries left; file doesn't exist

stage2notfound: db "Could not find YAKLDR.SYS", 0
SRCHFAILURE:
    mov     si, stage2notfound
    call    Print

    jmp     halt

stage2found: db "Found YAKLDR.SYS!", 0
LOAD_FAT:
    mov     si, stage2found
    call    Print

halt:
    cli
    hlt