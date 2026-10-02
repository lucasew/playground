#include "video.h"

#include "port.h"
#include "string.h"

#include <stdint.h>

static uint8_t *const fb = (uint8_t *)0xA0000;

struct bapl_header {
    uint32_t magic;
    uint16_t width;
    uint16_t height;
    uint16_t fps;
    uint16_t scale;
    uint32_t frame_count;
    uint32_t reserved;
} __attribute__((packed));

static const struct bapl_header *hdr;
static const uint32_t *offsets;
static const uint8_t *file;

void video_clear(void) {
    memset(fb, 0, 320u * 200u);
}

void video_init(void) {
    outb(0x3C8, 0);
    for (int i = 0; i < 256; i++) {
        uint8_t level = i == 0 ? 0 : 63;
        outb(0x3C9, level);
        outb(0x3C9, level);
        outb(0x3C9, level);
    }
    video_clear();
}

int video_open(const uint8_t *blob) {
    hdr = (const struct bapl_header *)blob;
    if (hdr->magic != 0x4C504142u || hdr->width == 0 || hdr->height == 0 || hdr->fps == 0) {
        return 0;
    }
    if (hdr->scale != 1 && hdr->scale != 2) {
        return 0;
    }
    if ((uint32_t)hdr->width * hdr->scale > 320 || (uint32_t)hdr->height * hdr->scale > 200) {
        return 0;
    }
    file = blob;
    offsets = (const uint32_t *)(blob + sizeof *hdr);
    return 1;
}

uint32_t video_frame_count(void) {
    return hdr->frame_count;
}

uint16_t video_fps(void) {
    return hdr->fps;
}

static void span(uint32_t x, uint32_t y, uint32_t n, uint8_t color) {
    uint32_t scale = hdr->scale;
    uint32_t x0 = (320u - (uint32_t)hdr->width * scale) / 2u;
    uint32_t y0 = (200u - (uint32_t)hdr->height * scale) / 2u;
    if (scale == 1) {
        memset(fb + (y0 + y) * 320u + x0 + x, color, n);
        return;
    }
    uint8_t *row0 = fb + (y0 + y * 2u) * 320u + x0 + x * 2u;
    uint8_t *row1 = row0 + 320u;
    memset(row0, color, n * 2u);
    memset(row1, color, n * 2u);
}

void video_blit(uint32_t index) {
    if (index >= hdr->frame_count) {
        index = hdr->frame_count - 1;
    }
    const uint8_t *p = file + offsets[index];
    const uint8_t *end = file + offsets[index + 1];
    uint32_t x = 0;
    uint32_t y = 0;
    uint32_t width = hdr->width;
    uint32_t height = hdr->height;

    while (p < end && y < height) {
        uint8_t packed = *p++;
        uint8_t color = packed >> 7;
        uint32_t len = (packed & 0x7Fu) + 1u;
        while (len && y < height) {
            uint32_t room = width - x;
            uint32_t n = len < room ? len : room;
            span(x, y, n, color);
            x += n;
            len -= n;
            if (x >= width) {
                x = 0;
                y++;
            }
        }
    }
    while (y < height) {
        span(x, y, width - x, 0);
        x = 0;
        y++;
    }
}
