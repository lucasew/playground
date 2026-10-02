#include "idt.h"
#include "music.h"
#include "port.h"
#include "rtc.h"
#include "serial.h"
#include "speaker.h"
#include "video.h"

#include <stdint.h>

extern uint8_t _binary_frames_rle_start[];
extern uint8_t _binary_music_bin_start[];

static void fail(const char *message) {
    serial_puts(message);
    disable_interrupts();
    for (;;) {
        halt();
    }
}

void kmain(void) {
    serial_init();
    serial_puts("\nchimia os\n");

    if (!video_open(_binary_frames_rle_start)) {
        fail("bad frames\n");
    }
    if (!music_open(_binary_music_bin_start)) {
        fail("bad music\n");
    }

    video_init();
    idt_init();
    pic_init();
    rtc_init();
    speaker_set(0);
    enable_interrupts();
    /* Position zero is this RTC sample. Later ticks subtract it. */
    rtc_sync();

    uint32_t frames = video_frame_count();
    uint32_t fps = video_fps();
    uint32_t song_ms = music_song_ms();
    /* The lead is silent until its first note. Stay on a blank screen
     * through the RTC latch, then start the picture at that note. */
    uint32_t lead = music_lead_ms();
    if (lead >= song_ms) {
        lead = 0;
    }
    uint32_t span = song_ms - lead;
    if (span == 0) {
        span = song_ms;
    }

    serial_puts("frames ");
    serial_putu(frames);
    serial_puts(" fps ");
    serial_putu(fps);
    serial_puts(" notes ");
    serial_putu(music_note_count());
    serial_puts("\nt0 ");
    serial_putu(rtc_initial_sod());
    serial_puts(" lead ");
    serial_putu(lead);
    serial_puts("\nplay\n");

    uint32_t shown = 0xFFFFFFFFu;
    uint32_t next_log = 0;
    for (;;) {
        /* One tick, one snapshot. Frame and note both come from it, and a
         * late blit just presents the latest snapshot instead of catching
         * up one missed tick at a time. */
        uint32_t elapsed = rtc_elapsed_ms();
        uint32_t pos = lead + (elapsed % span);
        uint32_t frame = (uint32_t)(((uint64_t)pos * fps) / 1000u);
        uint16_t hz = music_freq_at(pos);
        if (frame >= frames) {
            frame = frames - 1u;
        }
        if (frame != shown) {
            video_blit(frame);
            shown = frame;
        }
        speaker_set(hz);

        if (elapsed >= next_log) {
            serial_puts("t=");
            serial_putu(pos);
            serial_puts(" frame=");
            serial_putu(frame);
            serial_puts(" hz=");
            serial_putu(hz);
            serial_puts(" elapsed=");
            serial_putu(elapsed);
            serial_puts(" edges=");
            serial_putu(speaker_edges());
            serial_putc('\n');
            next_log = elapsed + 500u;
        }
        halt();
    }
}
