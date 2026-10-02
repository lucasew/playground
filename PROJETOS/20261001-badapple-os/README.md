# Chimia

A from-scratch x86_64 kernel that boots from its own BIOS loader, switches to long mode, and plays **Bad Apple!!** on the PC speaker while drawing the silhouette frames in VGA mode 13h (320×200).

The speaker can only make one square wave at a time, so it plays the vocal melody with the note starts and lengths from that MIDI part. The synth hook is a two-bar riff repeated for the whole song, so it is not used. The kernel latches the RTC before the first frame, then starts the picture on the melody's first note, so the silent intro is not shown. Frames are 1-bit, 15 fps, run-length encoded. Playback loops over that melody.

The show advances on the RTC periodic tick (1024 Hz). At boot the kernel stores one CMOS time; every tick, the frame and the note are both computed from how much later the RTC is than that sample. The tick count is only the fraction inside the current RTC second, so a missed interrupt cannot pull the picture and the buzzer apart.

Bad Apple!! is a Touhou fan recording (original theme by ZUN, arrangement associated with Alstroemeria Records / nomico, shadow PV by Anira). The media in `assets/` is only here so this personal demo can be rebuilt.

## Layout

- `boot/boot.asm` loads the second stage from disk.
- `boot/stage2.asm` sets mode 13h, reads the kernel to 1 MiB, and enters 64-bit mode.
- `kernel/` is freestanding C. IRQ8 is the RTC tick. Channel 2 of the PIT drives the speaker. Playback position is the difference from the initial RTC reading.
- `tools/prepare_media.py` turns the video and MIDI into `build/frames.rle` and `build/music.bin`.

`chimiaos/` in the workspace root is not writable from this environment, so the tree lives in `chimia/`.

## Build

Requires `nasm`, `gcc`, `ld`, `python3`, `ffmpeg`, and `qemu-system-x86_64`.

```sh
make
make smoke   # headless boot: two screenshots and a speaker WAV
make run     # window plus PC speaker, serial log on the terminal
```

`make run` uses SDL audio. The headless check records the speaker with QEMU's WAV backend instead:

```sh
qemu-system-x86_64 \
  -m 128 -no-reboot -net none -boot order=c,menu=off \
  -drive file=build/chimia.img,format=raw,if=ide,index=0,media=disk \
  -audiodev sdl,id=speaker -machine pc,pcspk-audiodev=speaker \
  -serial stdio
```

The bottom of the screen shows a white load bar until the kernel starts. Serial output looks like `B2V....DLK`, then `chimia os` and a `t= frame= hz= edges=` line twice a second.
