#include "music.h"

#include <stdint.h>

struct note_header {
    uint32_t magic;
    uint32_t count;
    uint32_t song_ms;
    uint32_t reserved;
} __attribute__((packed));

static const struct note_header *hdr;
static const struct note_event *notes;

int music_open(const uint8_t *blob) {
    hdr = (const struct note_header *)blob;
    if (hdr->magic != 0x45544F4Eu || hdr->song_ms == 0) {
        return 0;
    }
    notes = (const struct note_event *)(blob + sizeof *hdr);
    return 1;
}

uint32_t music_song_ms(void) {
    return hdr ? hdr->song_ms : 0;
}

uint32_t music_note_count(void) {
    return hdr->count;
}

uint32_t music_lead_ms(void) {
    if (!hdr || hdr->count == 0) {
        return 0;
    }
    return notes[0].start_ms;
}

uint16_t music_freq_at(uint32_t ms) {
    if (!hdr || hdr->count == 0) {
        return 0;
    }
    uint32_t count = hdr->count;
    uint32_t lo = 0;
    uint32_t hi = count;
    while (lo < hi) {
        uint32_t mid = lo + (hi - lo) / 2u;
        if (notes[mid].start_ms <= ms) {
            lo = mid + 1u;
        } else {
            hi = mid;
        }
    }
    if (lo == 0) {
        return 0;
    }
    const struct note_event *event = &notes[lo - 1u];
    if (ms < event->start_ms + event->dur_ms) {
        return event->freq;
    }
    return 0;
}
