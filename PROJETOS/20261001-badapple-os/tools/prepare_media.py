#!/usr/bin/env python3
"""Build 1-bit Bad Apple frames and a monophonic PC-speaker arrangement."""

import struct
import subprocess
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "assets"
BUILD = ROOT / "build"
VIDEO = ASSETS / "BadApple.mp4"
MIDI = ASSETS / "badapple-full.mid"

WIDTH = 320
HEIGHT = 200
FPS = 15
SCALE = 1
THRESHOLD = 128
# This encode of the PV is silent until the downbeat.
AUDIO_OFFSET_MS = 1330
DRUM_MS = 50
BAPL_MAGIC = 0x4C504142
NOTE_MAGIC = 0x45544F4E


def midi_hz(note):
    return int(round(440.0 * (2.0 ** ((note - 69) / 12.0))))


def drum_hz(note):
    if note in (35, 36):
        return 120
    if note in (38, 40):
        return 190
    if note in (41, 43, 45, 47, 48, 50):
        return 150
    if note in (42, 44, 46):
        return 880
    if note in (49, 51, 52, 55, 57, 59):
        return 640
    return 300


def read_varlen(data, index):
    value = 0
    while True:
        byte = data[index]
        index += 1
        value = (value << 7) | (byte & 0x7F)
        if byte < 0x80:
            return value, index


def parse_midi(data):
    if data[:4] != b"MThd":
        raise SystemExit("not a standard MIDI file")
    header_len = struct.unpack(">I", data[4:8])[0]
    fmt, ntracks, division = struct.unpack(">HHH", data[8:14])
    if fmt == 2 or division & 0x8000:
        raise SystemExit("unsupported MIDI time division")
    tracks = []
    index = 8 + header_len
    for _ in range(ntracks):
        if data[index:index + 4] != b"MTrk":
            raise SystemExit("truncated MIDI track")
        length = struct.unpack(">I", data[index + 4:index + 8])[0]
        index += 8
        end = index + length
        status = None
        tick = 0
        name = ""
        events = []
        while index < end:
            delta, index = read_varlen(data, index)
            tick += delta
            if index >= end:
                break
            byte = data[index]
            if byte >= 0x80:
                status = byte
                index += 1
            if status is None:
                break
            if status == 0xFF:
                meta = data[index]
                index += 1
                size, index = read_varlen(data, index)
                payload = data[index:index + size]
                index += size
                if meta == 0x03:
                    name = payload.decode("latin1", "replace")
                elif meta == 0x51 and len(payload) == 3:
                    tempo = (payload[0] << 16) | (payload[1] << 8) | payload[2]
                    events.append((tick, "tempo", tempo))
                if meta == 0x2F:
                    break
                continue
            if status in (0xF0, 0xF7):
                size, index = read_varlen(data, index)
                index += size
                continue
            kind = status & 0xF0
            channel = status & 0x0F
            nbytes = 2 if kind in (0x80, 0x90, 0xA0, 0xB0, 0xE0) else 1
            payload = data[index:index + nbytes]
            index += nbytes
            if kind == 0x90 and payload[1] > 0:
                events.append((tick, "on", channel, payload[0], payload[1]))
            elif kind == 0x80 or (kind == 0x90 and payload[1] == 0):
                events.append((tick, "off", channel, payload[0], 0))
        tracks.append({"name": name, "events": events})
    return division, tracks


def tempo_map(tracks):
    by_tick = {0: 500000}
    for track in tracks:
        for event in track["events"]:
            if event[1] == "tempo":
                by_tick[event[0]] = event[2]
    return sorted(by_tick.items())


def make_tick_to_ms(division, tempos):
    def tick_to_ms(tick):
        microseconds = 0.0
        last_tick = 0
        last_tempo = tempos[0][1]
        for point, tempo in tempos:
            if point >= tick:
                break
            if point > last_tick:
                microseconds += (point - last_tick) * last_tempo / division
            last_tick = point
            last_tempo = tempo
        microseconds += (tick - last_tick) * last_tempo / division
        return microseconds / 1000.0

    return tick_to_ms


def note_intervals(track, tick_to_ms):
    pending = {}
    intervals = []
    for event in track["events"]:
        if event[1] == "on":
            _, _, channel, note, _velocity = event
            pending.setdefault((channel, note), []).append(event[0])
        elif event[1] == "off":
            _, _, channel, note, _velocity = event
            key = (channel, note)
            if pending.get(key):
                start = pending[key].pop(0)
                intervals.append((tick_to_ms(start), tick_to_ms(event[0]), note))
    intervals.sort()
    return intervals


def flatten(intervals, priority, hz_of):
    """Collapse one part to non-overlapping segments, highest note wins."""
    points = []
    for start, end, note in intervals:
        if end > start:
            points.append((start, 1, note))
            points.append((end, 0, note))
    points.sort()
    active = Counter()
    segments = []
    current = None
    prev = None
    index = 0
    while index < len(points):
        time = points[index][0]
        if current is not None and prev is not None and time > prev:
            segments.append((prev, time, hz_of(current), priority))
        while index < len(points) and points[index][0] == time:
            _time, is_start, note = points[index]
            if is_start:
                active[note] += 1
            else:
                active[note] -= 1
                if active[note] <= 0:
                    del active[note]
            index += 1
        current = max(active) if active else None
        prev = time
    return segments


def priority_mix(layers):
    points = []
    for start, end, freq, priority in layers:
        if end > start:
            points.append((start, 1, priority, freq))
            points.append((end, 0, priority, freq))
    points.sort()
    active = {}
    segments = []
    current = 0
    prev = None
    index = 0

    def chosen():
        if not active:
            return 0
        priority = max(active)
        return max(active[priority])

    while index < len(points):
        time = points[index][0]
        if current and prev is not None and time > prev:
            segments.append((prev, time, current))
        while index < len(points) and points[index][0] == time:
            _time, is_start, priority, freq = points[index]
            bucket = active.setdefault(priority, Counter())
            if is_start:
                bucket[freq] += 1
            else:
                bucket[freq] -= 1
                if bucket[freq] <= 0:
                    del bucket[freq]
                if not bucket:
                    del active[priority]
            index += 1
        current = chosen()
        prev = time
    return segments


def shift_clip(segments, offset_ms, song_ms):
    shifted = []
    for start, end, freq in segments:
        start += offset_ms
        end += offset_ms
        if end <= 0 or start >= song_ms:
            continue
        start = max(0.0, start)
        end = min(float(song_ms), end)
        if end - start >= 15:
            shifted.append((start, end, freq))
    return shifted


def add_drum_clicks(melody, drums, song_ms):
    sounding = [(s, e) for s, e, _freq in melody]
    clicks = []
    for start, _end, note in drums:
        click_end = start + DRUM_MS
        if click_end > song_ms or start < 0:
            continue
        gap_start = 0.0
        blocked = False
        for seg_start, seg_end in sounding:
            if seg_end <= start:
                gap_start = max(gap_start, seg_end)
                continue
            if seg_start >= click_end:
                if seg_start - gap_start < 80 or click_end > seg_start:
                    blocked = True
                break
            blocked = True
            break
        else:
            if song_ms - gap_start < 80:
                blocked = True
        if not blocked and start >= gap_start and click_end - start >= 25:
            clicks.append((start, click_end, drum_hz(note)))
    return melody + clicks


def merge_segments(segments):
    segments = sorted(segments)
    merged = []
    for start, end, freq in segments:
        if not merged:
            merged.append([start, end, freq])
            continue
        prev_start, prev_end, prev_freq = merged[-1]
        if start < prev_end:
            if start <= prev_start:
                if freq >= prev_freq:
                    merged[-1] = [start, end, freq]
                continue
            merged[-1][1] = start
            if end > start:
                merged.append([start, end, freq])
            continue
        if freq == prev_freq and start - prev_end <= 8:
            merged[-1][1] = max(prev_end, end)
            continue
        merged.append([start, end, freq])
    segments = [(s, e, f) for s, e, f in merged]
    output = []
    for start, end, freq in merged:
        cursor = start
        while end - cursor > 60000:
            output.append((cursor, cursor + 60000, freq))
            cursor += 60000
        if end - cursor >= 20:
            output.append((cursor, end, freq))
    return output


def encode_frames():
    command = [
        "ffmpeg", "-v", "error", "-i", str(VIDEO),
        "-vf", f"fps={FPS},scale={WIDTH}:{HEIGHT}:flags=area,format=gray",
        "-f", "rawvideo", "-",
    ]
    process = subprocess.Popen(command, stdout=subprocess.PIPE)
    frame_bytes = WIDTH * HEIGHT
    payload = bytearray()
    offsets = []
    white_per_second = []
    white_acc = 0
    white_n = 0
    frames = 0
    while True:
        blob = process.stdout.read(frame_bytes)
        if len(blob) < frame_bytes:
            break
        offsets.append(len(payload))
        color = 1 if blob[0] >= THRESHOLD else 0
        run = 0
        whites = 0
        for pixel in blob:
            bit = 1 if pixel >= THRESHOLD else 0
            whites += bit
            if bit == color and run < 128:
                run += 1
                continue
            payload.append(((color & 1) << 7) | (run - 1))
            color = bit
            run = 1
        if run:
            payload.append(((color & 1) << 7) | (run - 1))
        white_acc += whites
        white_n += 1
        if white_n == FPS:
            white_per_second.append(white_acc / (white_n * frame_bytes))
            white_acc = 0
            white_n = 0
        frames += 1
        if frames % 300 == 0:
            print(f"  encoded {frames} frames", file=sys.stderr)
    process.stdout.close()
    if process.wait() != 0:
        raise SystemExit("ffmpeg failed while extracting frames")
    if white_n:
        white_per_second.append(white_acc / (white_n * frame_bytes))
    offsets.append(len(payload))
    if frames == 0:
        raise SystemExit("no frames extracted")
    header = struct.pack("<IHHHHII", BAPL_MAGIC, WIDTH, HEIGHT, FPS, SCALE, frames, 0)
    table = b"".join(struct.pack("<I", 20 + (frames + 1) * 4 + off) for off in offsets)
    # offsets currently point at payload-relative positions. Rewrite absolute.
    base = 20 + (frames + 1) * 4
    table = b"".join(struct.pack("<I", base + off) for off in offsets)
    blob = header + table + payload
    return blob, frames, white_per_second


def speaker_hz(freq):
    freq = int(freq)
    while 0 < freq < 55:
        freq *= 2
    return freq


def articulate(melody, drums, song_ms, tempo_us, offset_ms):
    """Pulse every sixteenth note.

    The animation cuts on the 16th-note grid. Legato MIDI notes hold the
    speaker across several of those cuts, so the tune sounds much slower
    than the picture. Short gaps force the PC speaker to re-attack.
    """
    step_us = max(int(tempo_us) // 4, 50000)
    gap_us = max(step_us // 6, 12000)
    on_us = step_us - gap_us
    melody = sorted(melody)
    drums = sorted(drums)
    melody_index = 0
    drum_index = 0
    notes = []
    start_us = int(offset_ms) * 1000
    limit_us = int(song_ms) * 1000
    while start_us < limit_us:
        start = start_us / 1000.0
        slot_end = start_us + step_us
        slot_end_ms = slot_end / 1000.0
        while melody_index < len(melody) and melody[melody_index][1] <= start:
            melody_index += 1
        freq = 0
        # Any melodic note that overlaps the slot wins. Sampling only the
        # slot start missed notes that begin a millisecond later, and a
        # drum click replaced the lead.
        if (
            melody_index < len(melody)
            and melody[melody_index][0] < slot_end_ms
            and melody[melody_index][1] > start
        ):
            freq = speaker_hz(melody[melody_index][2])
        if not freq:
            while drum_index < len(drums) and drums[drum_index][0] < start:
                drum_index += 1
            if drum_index < len(drums) and drums[drum_index][0] < slot_end_ms:
                freq = speaker_hz(drum_hz(drums[drum_index][2]))
                while drum_index < len(drums) and drums[drum_index][0] < slot_end_ms:
                    drum_index += 1
        if freq:
            notes.append((start, min(float(song_ms), (start_us + on_us) / 1000.0), freq))
        start_us = slot_end
    return notes


def find_track(tracks, prefix):
    prefix = prefix.lower()
    for track in tracks:
        if track["name"].lower().startswith(prefix):
            return track
    raise SystemExit(f"MIDI is missing a {prefix!r} track")


def build_music(frame_count):
    division, tracks = parse_midi(MIDI.read_bytes())
    tempos = tempo_map(tracks)
    tick_to_ms = make_tick_to_ms(division, tempos)
    song_ms = frame_count * 1000 // FPS
    # The vocal line is the melody. Synth1 is a two-bar hook repeated for
    # the whole song, so playing it alone sounds like a few seconds looped.
    # Tied sixteenths of the same pitch are one sung note.
    vocal = note_intervals(find_track(tracks, "vocals"), tick_to_ms)
    melody = [(start, end, freq) for start, end, freq, _priority in flatten(vocal, 1, midi_hz)]
    melody = shift_clip(melody, AUDIO_OFFSET_MS, song_ms)
    melody = merge_segments(melody)
    timed = []
    for start, end, freq in melody:
        start_i = int(round(start))
        end_i = int(round(end))
        if end_i - start_i < 20:
            continue
        timed.append([start_i, end_i, int(freq)])
    for index in range(1, len(timed)):
        if timed[index][0] < timed[index - 1][1]:
            timed[index - 1][1] = timed[index][0]
    timed = [item for item in timed if item[1] - item[0] >= 20]
    notes = bytearray()
    notes += struct.pack("<IIII", NOTE_MAGIC, len(timed), song_ms, AUDIO_OFFSET_MS)
    for start, end, freq in timed:
        duration = min(end - start, 65535)
        notes += struct.pack("<IHH", start, freq, duration)
    return notes, song_ms, len(timed), tempos[0][1]


def main():
    if not VIDEO.exists() or not MIDI.exists():
        raise SystemExit(f"missing {VIDEO} or {MIDI}")
    BUILD.mkdir(parents=True, exist_ok=True)
    print("extracting frames", file=sys.stderr)
    frames, count, white = encode_frames()
    print("arranging buzzer part", file=sys.stderr)
    music, song_ms, note_count, tempo = build_music(count)
    (BUILD / "frames.rle").write_bytes(frames)
    (BUILD / "music.bin").write_bytes(music)
    manifest = [f"frames={count}", f"fps={FPS}", f"song_ms={song_ms}", f"notes={note_count}", f"tempo_us={tempo}"]
    for second, ratio in enumerate(white):
        manifest.append(f"white {second} {ratio:.4f}")
    (BUILD / "manifest.txt").write_text("\n".join(manifest) + "\n")
    print(f"frames {count}  rle {len(frames)} bytes  notes {note_count}  song {song_ms} ms")


if __name__ == "__main__":
    main()
