# Linux amd64 port. This is the whole operating-system surface.
# hal_write(ptr, len) writes bytes to the standard error stream.
# hal_exit(code) ends the process.

.intel_syntax noprefix
.text

.global _start
_start:
    xor rbp, rbp
    call rt_start
    mov rdi, rax
    call hal_exit

.global hal_write
hal_write:
    mov rdx, rsi
    mov rsi, rdi
    mov rdi, 2
    mov rax, 1
    syscall
    ret

.global hal_exit
hal_exit:
    mov rax, 60
    syscall

.section .note.GNU-stack,"",@progbits
