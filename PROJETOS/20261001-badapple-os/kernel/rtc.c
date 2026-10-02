#include "rtc.h"

#include "port.h"

#include <stdint.h>

/* Register A rate select 6: 32768 >> 5 = 1024 Hz, from the RTC crystal. */
#define RTC_RATE 6

static volatile uint32_t rtc_ticks;
static volatile uint32_t clock_ready;
static uint32_t tick_base;
static uint32_t sec_since;
static int initial_sod = -1;
static uint8_t last_sec = 0xFF;
static uint8_t last_min = 0xFF;
static uint8_t last_hour = 0xFF;

static uint8_t cmos_read(uint8_t reg) {
    outb(0x70, (uint8_t)(reg | 0x80));
    io_wait();
    return inb(0x71);
}

static void cmos_write(uint8_t reg, uint8_t value) {
    outb(0x70, (uint8_t)(reg | 0x80));
    io_wait();
    outb(0x71, value);
    io_wait();
}

static int from_bcd(uint8_t value) {
    return (value >> 4) * 10 + (value & 0x0F);
}

struct civil {
    int hour;
    int min;
    int sec;
};

static int sod_of(const struct civil *time) {
    return (time->hour * 60 + time->min) * 60 + time->sec;
}

static int read_civil(struct civil *out) {
    disable_interrupts();
    for (int attempt = 0; attempt < 8; attempt++) {
        int guard = 0;
        while ((cmos_read(0x0A) & 0x80) && guard++ < 100000) {
        }
        uint8_t sec1 = cmos_read(0x00);
        uint8_t min1 = cmos_read(0x02);
        uint8_t hour1 = cmos_read(0x04);
        uint8_t status_b = cmos_read(0x0B);
        guard = 0;
        while ((cmos_read(0x0A) & 0x80) && guard++ < 100000) {
        }
        uint8_t sec2 = cmos_read(0x00);
        uint8_t min2 = cmos_read(0x02);
        uint8_t hour2 = cmos_read(0x04);
        if (sec1 != sec2 || min1 != min2 || hour1 != hour2) {
            continue;
        }

        int binary = (status_b & 0x04) != 0;
        int hour24 = (status_b & 0x02) != 0;
        int sec = binary ? sec1 : from_bcd(sec1);
        int min = binary ? min1 : from_bcd(min1);
        uint8_t hour_raw = hour1;
        int pm = 0;
        if (!hour24) {
            pm = (hour_raw & 0x80) != 0;
            hour_raw = (uint8_t)(hour_raw & 0x7F);
        }
        int hour = binary ? hour_raw : from_bcd(hour_raw);
        if (!hour24) {
            if (hour == 12) {
                hour = 0;
            }
            if (pm) {
                hour += 12;
            }
        }
        enable_interrupts();
        if (sec > 59 || min > 59 || hour > 23) {
            return 0;
        }
        out->hour = hour;
        out->min = min;
        out->sec = sec;
        return 1;
    }
    enable_interrupts();
    return 0;
}

static int seconds_between(int now, int start) {
    int delta = now - start;
    if (delta < 0) {
        delta += 86400;
    }
    return delta;
}

void rtc_init(void) {
    uint8_t status_a = cmos_read(0x0A);
    uint8_t divider = (uint8_t)(status_a & 0x70);
    if (divider == 0) {
        divider = 0x20;
    }
    cmos_write(0x0A, (uint8_t)(divider | RTC_RATE));

    uint8_t status_b = cmos_read(0x0B);
    cmos_write(0x0B, (uint8_t)((status_b | 0x40) & (uint8_t)~0x80));
    (void)cmos_read(0x0C);
}

void rtc_irq(void) {
    (void)cmos_read(0x0C);
    rtc_ticks++;
}

void rtc_sync(void) {
    struct civil first;
    struct civil now;
    if (!read_civil(&first)) {
        first.hour = 0;
        first.min = 0;
        first.sec = 0;
        now = first;
    } else {
        now = first;
        for (uint32_t spin = 0; spin < 4000000u && now.sec == first.sec; spin++) {
            if (!read_civil(&now)) {
                now = first;
            }
        }
    }
    initial_sod = sod_of(&now);
    sec_since = 0;
    last_hour = (uint8_t)now.hour;
    last_min = (uint8_t)now.min;
    last_sec = (uint8_t)now.sec;
    disable_interrupts();
    tick_base = rtc_ticks;
    clock_ready = 1;
    enable_interrupts();
}

uint32_t rtc_elapsed_ms(void) {
    struct civil now;
    int got = read_civil(&now);
    disable_interrupts();
    if (got && clock_ready &&
        (now.sec != last_sec || now.min != last_min || now.hour != last_hour)) {
        sec_since = (uint32_t)seconds_between(sod_of(&now), initial_sod);
        tick_base = rtc_ticks;
        last_hour = (uint8_t)now.hour;
        last_min = (uint8_t)now.min;
        last_sec = (uint8_t)now.sec;
    }
    uint32_t into = rtc_ticks - tick_base;
    uint32_t secs = sec_since;
    enable_interrupts();
    /* The crystal second is the authority. Ticks may not run past it. */
    if (into > 1024u) {
        into = 1024u;
    }
    uint32_t into_ms = (into * 1000u) / 1024u;
    if (into_ms > 999u) {
        into_ms = 999u;
    }
    return secs * 1000u + into_ms;
}

uint32_t rtc_initial_sod(void) {
    return initial_sod < 0 ? 0u : (uint32_t)initial_sod;
}
