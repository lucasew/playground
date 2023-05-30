; vim:ft=asm

segment .data
virgula db ", ", 0
nl db 10, 0
fmt db "%i", 0

vetor dq 1, 2, 3, 4, 5, 6, 7, 8, 9, 0

segment .text

extern printf
global main


main:
    mov r12, 0 ; indice
    jmp _body
_virgula:
    push rbp
    mov rdi, virgula
    ; xor rsi, rsi
    mov rax, 0
    call printf
    pop rbp
_body:
    mov rsi, [vetor+r12*8]
    push rbp
    mov rdi, fmt
    mov rax, 0
    call printf
    pop rbp

    add r12, 1
    mov rsi, [vetor+r12*8]
    cmp rsi, 0
    je _end

    jmp _virgula

_end:
    push rbp
    mov rdi, nl
    mov rax, 0
    call printf
    pop rbp

    mov rax, 0
    ret
