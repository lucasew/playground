# Cooperative actor runtime. It calls hal_write and hal_exit only.
# Actors are rows of eight integer slots. A pid call enqueues a method id.
# rt_await runs queued methods until that promise has a value.

.intel_syntax noprefix
.text

.global rt_start
rt_start:
    call lang_main
    xor rax, rax
    ret

.global rt_actor
rt_actor:
    mov rax, qword ptr [rip+next_actor]
    lea rcx, [rax+1]
    mov qword ptr [rip+next_actor], rcx
    mov rdi, rax
    shl rdi, 6
    lea rsi, [rip+actors]
    add rdi, rsi
    xor ecx, ecx
.Lzero:
    mov qword ptr [rdi+rcx*8], 0
    inc ecx
    cmp ecx, 8
    jb .Lzero
    ret

.global rt_load
rt_load:
    mov rax, rdi
    shl rax, 6
    lea rcx, [rip+actors]
    add rcx, rax
    mov rax, qword ptr [rcx+rsi*8]
    ret

.global rt_store
rt_store:
    mov rax, rdi
    shl rax, 6
    lea rcx, [rip+actors]
    add rcx, rax
    mov qword ptr [rcx+rsi*8], rdx
    ret

.global rt_call
rt_call:
    mov rax, qword ptr [rip+next_promise]
    lea rcx, [rax+1]
    mov qword ptr [rip+next_promise], rcx
    mov r8, qword ptr [rip+mbox_w]
    imul r9, r8, 24
    lea r10, [rip+mbox]
    mov qword ptr [r10+r9], rdi
    mov qword ptr [r10+r9+8], rsi
    mov qword ptr [r10+r9+16], rax
    inc r8
    mov qword ptr [rip+mbox_w], r8
    lea r10, [rip+prom_ready]
    mov byte ptr [r10+rax], 0
    ret

.global rt_await
rt_await:
    push rbp
    mov rbp, rsp
    push rbx
    push r12
    push r13
    push r14
    push r15
    sub rsp, 8
    mov rbx, rdi
.Lawait:
    lea rcx, [rip+prom_ready]
    cmp byte ptr [rcx+rbx], 0
    jne .Lready
    mov r12, qword ptr [rip+mbox_r]
    cmp r12, qword ptr [rip+mbox_w]
    jae .Ldeadlock
    imul r11, r12, 24
    lea rcx, [rip+mbox]
    mov r13, qword ptr [rcx+r11]
    mov r14, qword ptr [rcx+r11+8]
    mov r15, qword ptr [rcx+r11+16]
    inc r12
    mov qword ptr [rip+mbox_r], r12
    lea rax, [rip+lang_methods]
    mov rax, qword ptr [rax+r14*8]
    mov rdi, r13
    call rax
    lea rcx, [rip+prom_val]
    mov qword ptr [rcx+r15*8], rax
    lea rcx, [rip+prom_ready]
    mov byte ptr [rcx+r15], 1
    jmp .Lawait
.Lready:
    lea rcx, [rip+prom_val]
    mov rax, qword ptr [rcx+rbx*8]
    add rsp, 8
    pop r15
    pop r14
    pop r13
    pop r12
    pop rbx
    pop rbp
    ret
.Ldeadlock:
    mov rdi, 2
    call hal_exit

.global rt_write
rt_write:
    jmp hal_write

.global rt_write_i64
rt_write_i64:
    push rbp
    mov rbp, rsp
    sub rsp, 32
    mov rax, rdi
    lea rcx, [rbp-1]
    mov r8, 10
    cmp rax, 0
    jne .Ldigits
    mov byte ptr [rcx], '0'
    mov rdi, rcx
    mov rsi, 1
    call rt_write
    jmp .Lnumdone
.Ldigits:
    xor rdx, rdx
    div r8
    add rdx, '0'
    mov byte ptr [rcx], dl
    dec rcx
    cmp rax, 0
    jne .Ldigits
    inc rcx
    mov rdi, rcx
    lea rsi, [rbp-1]
    sub rsi, rcx
    inc rsi
    call rt_write
.Lnumdone:
    leave
    ret

.bss
.align 8
actors:
    .zero 64*64
mbox:
    .zero 128*24
prom_val:
    .zero 256*8
prom_ready:
    .zero 256

.data
.align 8
next_actor:
    .quad 0
next_promise:
    .quad 0
mbox_r:
    .quad 0
mbox_w:
    .quad 0

.section .note.GNU-stack,"",@progbits
