#pragma once

#include <stdint.h>

/* The RTC periodic interrupt is the only tick. Playback position is the
 * time since the initial CMOS reading, not a counter that can miss IRQs:
 * whole seconds come from the clock, and the ticks only fill in the
 * current second. */
void rtc_init(void);
void rtc_sync(void);
void rtc_irq(void);
uint32_t rtc_elapsed_ms(void);
uint32_t rtc_initial_sod(void);
