; First sector. Loads stage2 from LBA 1 and jumps to it.
bits 16
org 0x7C00

stage2_sectors equ 16

start:
    cli
    xor ax, ax
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7C00
    mov [0x0500], dl
    call serial_init
    mov al, 'B'
    call serial_putc
    sti

    mov ah, 0x41
    mov bx, 0x55AA
    mov dl, [0x0500]
    int 0x13
    jc fail
    cmp bx, 0xAA55
    jne fail

    mov si, dap
    mov ah, 0x42
    mov dl, [0x0500]
    int 0x13
    jc fail

    jmp 0x0000:0x7E00

fail:
    mov al, 'E'
    call serial_putc
.hang:
    cli
    hlt
    jmp .hang

serial_init:
    mov dx, 0x3F9
    xor al, al
    out dx, al
    mov dx, 0x3FB
    mov al, 0x80
    out dx, al
    mov dx, 0x3F8
    mov al, 0x01
    out dx, al
    mov dx, 0x3F9
    xor al, al
    out dx, al
    mov dx, 0x3FB
    mov al, 0x03
    out dx, al
    mov dx, 0x3FA
    mov al, 0xC7
    out dx, al
    mov dx, 0x3FC
    mov al, 0x0B
    out dx, al
    ret

serial_putc:
    push dx
    push ax
    mov dx, 0x3FD
.wait:
    in al, dx
    test al, 0x20
    jz .wait
    pop ax
    push ax
    mov dx, 0x3F8
    out dx, al
    pop ax
    pop dx
    ret

align 4
dap:
    db 0x10
    db 0
    dw stage2_sectors
    dw 0x7E00
    dw 0x0000
    dq 1

times 510 - ($ - $$) db 0
dw 0xAA55
