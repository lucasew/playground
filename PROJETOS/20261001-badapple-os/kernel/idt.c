#include "idt.h"

#include "port.h"
#include "rtc.h"
#include "serial.h"

#include <stdint.h>

struct interrupt_frame {
    uint64_t rax, rcx, rdx, rbx, rbp, rsi, rdi;
    uint64_t r8, r9, r10, r11, r12, r13, r14, r15;
    uint64_t vector;
    uint64_t error;
    uint64_t rip, cs, rflags, rsp, ss;
};

struct idt_entry {
    uint16_t offset_low;
    uint16_t selector;
    uint8_t ist;
    uint8_t type_attr;
    uint16_t offset_mid;
    uint32_t offset_high;
    uint32_t zero;
} __attribute__((packed));

struct idt_ptr {
    uint16_t limit;
    uint64_t base;
} __attribute__((packed));

extern uint64_t isr_stub_table[];

static struct idt_entry idt[256];

static void set_gate(int vector, uint64_t handler) {
    idt[vector].offset_low = (uint16_t)handler;
    idt[vector].selector = 0x28;
    idt[vector].ist = 0;
    idt[vector].type_attr = 0x8E;
    idt[vector].offset_mid = (uint16_t)(handler >> 16);
    idt[vector].offset_high = (uint32_t)(handler >> 32);
    idt[vector].zero = 0;
}

void pic_init(void) {
    outb(0x20, 0x11);
    io_wait();
    outb(0xA0, 0x11);
    io_wait();
    outb(0x21, 0x20);
    io_wait();
    outb(0xA1, 0x28);
    io_wait();
    outb(0x21, 0x04);
    io_wait();
    outb(0xA1, 0x02);
    io_wait();
    outb(0x21, 0x01);
    io_wait();
    outb(0xA1, 0x01);
    io_wait();
    /* IRQ2 is the cascade. IRQ8 (slave line 0) is the RTC tick. */
    outb(0x21, 0xFB);
    outb(0xA1, 0xFE);
}

void idt_init(void) {
    for (int i = 0; i < 256; i++) {
        set_gate(i, isr_stub_table[i]);
    }
    struct idt_ptr ptr;
    ptr.limit = (uint16_t)(sizeof(idt) - 1);
    ptr.base = (uint64_t)&idt;
    __asm__ volatile("lidt %0" : : "m"(ptr));
}

void interrupt_dispatch(struct interrupt_frame *frame) {
    /* Vector 40 is IRQ8. Ack the RTC here; the main loop presents. */
    if (frame->vector == 40) {
        rtc_irq();
        outb(0xA0, 0x20);
        outb(0x20, 0x20);
        return;
    }
    if (frame->vector >= 32) {
        if (frame->vector >= 40) {
            outb(0xA0, 0x20);
        }
        outb(0x20, 0x20);
        return;
    }

    serial_puts("exc ");
    serial_putu((uint32_t)frame->vector);
    serial_puts(" rip ");
    serial_putu((uint32_t)frame->rip);
    serial_putc('\n');
    disable_interrupts();
    for (;;) {
        halt();
    }
}
