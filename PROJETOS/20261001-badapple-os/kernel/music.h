#pragma once

#include <stdint.h>

struct note_event {
    uint32_t start_ms;
    uint16_t freq;
    uint16_t dur_ms;
};

int music_open(const uint8_t *blob);
uint32_t music_song_ms(void);
uint32_t music_note_count(void);
/* Song time of the first note. The picture waits behind this. */
uint32_t music_lead_ms(void);
uint16_t music_freq_at(uint32_t ms);
