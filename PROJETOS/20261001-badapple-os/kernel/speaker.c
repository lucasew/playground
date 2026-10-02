#include "port.h"

#include <stdint.h>

static uint16_t programmed = 0xFFFF;
static uint32_t edges;

uint32_t speaker_edges(void) {
    return edges;
}

void speaker_set(uint16_t hz) {
    if (hz == programmed) {
        return;
    }
    programmed = hz;
    edges++;
    if (hz < 19) {
        outb(0x61, inb(0x61) & (uint8_t)~0x03);
        return;
    }
    uint32_t divisor = 1193182u / hz;
    outb(0x43, 0xB6);
    io_wait();
    outb(0x42, (uint8_t)(divisor & 0xFF));
    io_wait();
    outb(0x42, (uint8_t)((divisor >> 8) & 0xFF));
    io_wait();
    outb(0x61, inb(0x61) | 0x03);
}
