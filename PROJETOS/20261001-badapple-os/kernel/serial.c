#include "serial.h"
#include "port.h"

#define COM1 0x3F8

void serial_init(void) {
    outb(COM1 + 1, 0x00);
    outb(COM1 + 3, 0x80);
    outb(COM1 + 0, 0x01);
    outb(COM1 + 1, 0x00);
    outb(COM1 + 3, 0x03);
    outb(COM1 + 2, 0xC7);
    outb(COM1 + 4, 0x0B);
}

void serial_putc(char c) {
    while ((inb(COM1 + 5) & 0x20) == 0) {
    }
    outb(COM1, (uint8_t)c);
}

void serial_puts(const char *s) {
    while (*s) {
        if (*s == '\n') {
            serial_putc('\r');
        }
        serial_putc(*s++);
    }
}

void serial_putu(uint32_t value) {
    char buf[10];
    int i = 0;
    if (value == 0) {
        serial_putc('0');
        return;
    }
    while (value) {
        buf[i++] = (char)('0' + (value % 10));
        value /= 10;
    }
    while (i--) {
        serial_putc(buf[i]);
    }
}
