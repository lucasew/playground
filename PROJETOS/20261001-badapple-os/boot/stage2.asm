; Loaded at 0x7E00. Reads the kernel to 1 MiB, enables long mode, jumps to it.
; VGA mode 13h is set while the BIOS is still available.
bits 16
org 0x7E00

stage2:
    cli
    xor ax, ax
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7000
    cld
    call serial_init
    mov al, '2'
    call serial_putc

    ; Fast A20. Bit 0 of port 0x92 is system reset, so leave it clear.
    in al, 0x92
    or al, 00000010b
    and al, 11111110b
    out 0x92, al

    mov ax, 0x0013
    int 0x10
    xor ax, ax
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7000
    mov al, 'V'
    call serial_putc

    mov eax, [kernel_lba]
    mov [lba], eax
    mov dword [lba + 4], 0
    mov eax, [kernel_sectors]
    mov [sectors_left], eax
    mov [total_sectors], eax
    mov dword [dest], 0x100000
    test eax, eax
    jnz load_loop
    mov al, 'Z'
    call serial_putc
    jmp hang16

load_loop:
    cmp dword [sectors_left], 0
    je load_done

    mov eax, 32
    cmp eax, dword [sectors_left]
    jbe .count_ok
    mov eax, dword [sectors_left]
.count_ok:
    mov [chunk_sectors], eax
    mov [dap_count], ax

    mov eax, [lba]
    mov [dap_lba], eax
    mov eax, [lba + 4]
    mov [dap_lba + 4], eax

    mov byte [tries], 4
.retry:
    xor ax, ax
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7000
    sti
    mov si, dap
    mov ah, 0x42
    mov dl, [0x0500]
    int 0x13
    jnc .read_ok
    xor ax, ax
    mov ds, ax
    mov es, ax
    dec byte [tries]
    jz disk_fail
    mov dl, [0x0500]
    xor ah, ah
    int 0x13
    jmp .retry

.read_ok:
    xor ax, ax
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7000
    cli
    mov eax, [chunk_sectors]
    shl eax, 9
    mov [chunk_bytes], eax

    o32 lgdt [gdt_desc]
    mov eax, cr0
    or eax, 1
    mov cr0, eax
    jmp 0x08:copy32

; Returned to from protected mode after each chunk.
load_continue:
    mov eax, [chunk_sectors]
    add dword [lba], eax
    adc dword [lba + 4], 0
    sub dword [sectors_left], eax

    mov eax, [total_sectors]
    sub eax, dword [sectors_left]
    test eax, 0x3F
    jnz load_loop
    mov al, '.'
    call serial_putc
    jmp load_loop

load_done:
    mov al, 'D'
    call serial_putc
    cli
    o32 lgdt [gdt_desc]
    mov eax, cr0
    or eax, 1
    mov cr0, eax
    jmp 0x08:enter_long

disk_fail:
    mov al, 'X'
    call serial_putc
hang16:
    cli
.hang:
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
dap_count:
    dw 0
    dw 0
    dw 0x1000
dap_lba:
    dq 0

lba:            dq 0
dest:           dd 0
sectors_left:   dd 0
total_sectors:  dd 0
chunk_sectors:  dd 0
chunk_bytes:    dd 0
tries:          db 0

align 8
gdt:
    dq 0
    ; 0x08 32-bit code
    dw 0xFFFF
    dw 0x0000
    db 0x00
    db 10011010b
    db 11001111b
    db 0x00
    ; 0x10 32-bit data
    dw 0xFFFF
    dw 0x0000
    db 0x00
    db 10010010b
    db 11001111b
    db 0x00
    ; 0x18 16-bit code
    dw 0xFFFF
    dw 0x0000
    db 0x00
    db 10011010b
    db 00000000b
    db 0x00
    ; 0x20 16-bit data
    dw 0xFFFF
    dw 0x0000
    db 0x00
    db 10010010b
    db 00000000b
    db 0x00
    ; 0x28 64-bit code
    dw 0xFFFF
    dw 0x0000
    db 0x00
    db 10011010b
    db 10101111b
    db 0x00
    ; 0x30 64-bit data
    dw 0xFFFF
    dw 0x0000
    db 0x00
    db 10010010b
    db 00000000b
    db 0x00
gdt_end:

gdt_desc:
    dw gdt_end - gdt - 1
    dd gdt

; ---------------------------------------------------------------- 32-bit
bits 32
copy32:
    mov ax, 0x10
    mov ds, ax
    mov es, ax
    mov ss, ax
    cld
    mov esi, 0x10000
    mov edi, [dest]
    mov ecx, [chunk_bytes]
    rep movsb
    mov [dest], edi

    ; Progress bar along the bottom row. Color 15 is white in the BIOS palette.
    mov eax, [dest]
    sub eax, 0x100000
    shr eax, 9
    mov ebx, 320
    mul ebx
    mov ebx, [total_sectors]
    test ebx, ebx
    jz .no_bar
    div ebx
    cmp eax, 320
    jbe .bar_width
    mov eax, 320
.bar_width:
    mov ecx, eax
    push ecx
    mov edi, 0xA0000 + 198 * 320
    mov ecx, 320 * 2
    xor eax, eax
    rep stosb
    pop ecx
    test ecx, ecx
    jz .no_bar
    mov edi, 0xA0000 + 198 * 320
    mov al, 15
    rep stosb
.no_bar:

    jmp word 0x18:copy_to_real

bits 16
copy_to_real:
    mov ax, 0x20
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7000
    mov eax, cr0
    and eax, 0x7FFFFFFE
    mov cr0, eax
    jmp 0x0000:load_continue

bits 32
enter_long:
    mov ax, 0x10
    mov ds, ax
    mov es, ax
    mov ss, ax

    mov edi, 0x1000
    xor eax, eax
    mov ecx, 3072
    rep stosd

    mov dword [0x1000], 0x2003
    mov dword [0x2000], 0x3003
    mov edi, 0x3000
    mov eax, 0x83
    mov ecx, 512
.fill_pd:
    mov dword [edi], eax
    mov dword [edi + 4], 0
    add eax, 0x200000
    add edi, 8
    loop .fill_pd

    mov eax, cr4
    or eax, 1 << 5
    mov cr4, eax

    mov eax, 0x1000
    mov cr3, eax

    mov ecx, 0xC0000080
    rdmsr
    or eax, 1 << 8
    wrmsr

    mov eax, cr0
    or eax, 0x80000000
    mov cr0, eax

    jmp 0x28:long_mode

; ---------------------------------------------------------------- 64-bit
bits 64
long_mode:
    mov ax, 0x30
    mov ds, ax
    mov es, ax
    mov ss, ax
    xor ax, ax
    mov fs, ax
    mov gs, ax
    mov rsp, 0x90000
    mov al, 'L'
    mov dx, 0x3FD
.wait:
    in al, dx
    test al, 0x20
    jz .wait
    mov al, 'L'
    mov dx, 0x3F8
    out dx, al
    jmp 0x100000

; Footer patched by tools/mkimage.py. kernel_lba is a sector number.
bits 16
%if ($ - $$) > 8192 - 8
%error stage2 is larger than the 16 sectors reserved for it
%endif
times (8192 - 8) - ($ - $$) db 0
kernel_lba:     dd 17
kernel_sectors: dd 0
