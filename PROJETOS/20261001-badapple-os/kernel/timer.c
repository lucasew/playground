#include "port.h"

#include <stdint.h>

/* 1193182 / 1193 = 1000.15 Hz, close enough to treat one tick as one millisecond.
 * Mode 2 (rate generator). Mode 3 is a square wave; some boards then
 * deliver IRQ0 at half the rate QEMU does, which drags playback. */
volatile uint32_t timer_ticks;

void timer_init(void) {
    uint16_t divisor = 1193;
    outb(0x43, 0x34);
    io_wait();
    outb(0x40, (uint8_t)(divisor & 0xFF));
    io_wait();
    outb(0x40, (uint8_t)(divisor >> 8));
    io_wait();
}

uint32_t timer_ms(void) {
    return timer_ticks;
}
