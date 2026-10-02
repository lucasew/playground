bits 64
section .text

extern interrupt_dispatch

%macro define_stub 1
global isr_stub_%1
isr_stub_%1:
%if (%1 = 8) || (%1 = 10) || (%1 = 11) || (%1 = 12) || (%1 = 13) || (%1 = 14) || (%1 = 17) || (%1 = 21) || (%1 = 29) || (%1 = 30)
%else
    push qword 0
%endif
    push qword %1
    jmp isr_common
%endmacro

%assign vec 0
%rep 256
define_stub vec
%assign vec vec + 1
%endrep

global isr_stub_table
isr_stub_table:
%assign vec 0
%rep 256
    dq isr_stub_%+vec
%assign vec vec + 1
%endrep

isr_common:
    push r15
    push r14
    push r13
    push r12
    push r11
    push r10
    push r9
    push r8
    push rdi
    push rsi
    push rbp
    push rbx
    push rdx
    push rcx
    push rax
    mov rdi, rsp
    mov rbx, rsp
    and rsp, -16
    cld
    call interrupt_dispatch
    mov rsp, rbx
    pop rax
    pop rcx
    pop rdx
    pop rbx
    pop rbp
    pop rsi
    pop rdi
    pop r8
    pop r9
    pop r10
    pop r11
    pop r12
    pop r13
    pop r14
    pop r15
    add rsp, 16
    iretq
