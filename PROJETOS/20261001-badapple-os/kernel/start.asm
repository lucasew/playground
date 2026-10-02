bits 64
section .text.boot
global _start
extern kmain
extern bss_start
extern bss_end

_start:
    mov ax, 0x30
    mov ds, ax
    mov es, ax
    mov ss, ax
    xor ax, ax
    mov fs, ax
    mov gs, ax
    mov rsp, 0x90000
    cld

    mov dx, 0x3FD
.wait:
    in al, dx
    test al, 0x20
    jz .wait
    mov al, 'K'
    mov dx, 0x3F8
    out dx, al

    mov rdi, bss_start
    mov rcx, bss_end
    sub rcx, rdi
    xor eax, eax
    rep stosb

    and rsp, -16
    call kmain

.hang:
    cli
    hlt
    jmp .hang
