NASM ?= $(shell command -v nasm 2>/dev/null || echo /tmp/prefix/usr/bin/nasm)
CC ?= gcc
LD ?= ld
OBJCOPY ?= objcopy

BUILD := build
CFLAGS := -std=c11 -ffreestanding -fno-stack-protector -fno-pic -fno-pie \
	-mno-red-zone -mno-mmx -mno-sse -mno-sse2 -m64 -O2 -Wall -Wextra \
	-fno-asynchronous-unwind-tables -fno-unwind-tables \
	-mgeneral-regs-only -fcf-protection=none -I kernel

KERNEL_OBJS := $(BUILD)/start.o $(BUILD)/idt_asm.o $(BUILD)/main.o $(BUILD)/serial.o \
	$(BUILD)/string.o $(BUILD)/idt.o $(BUILD)/timer.o $(BUILD)/rtc.o $(BUILD)/speaker.o \
	$(BUILD)/video.o $(BUILD)/music.o $(BUILD)/frames.o $(BUILD)/musicbin.o

.PHONY: all clean smoke run media

all: $(BUILD)/chimia.img

media: $(BUILD)/media.stamp

$(BUILD)/media.stamp: tools/prepare_media.py assets/BadApple.mp4 assets/badapple-full.mid
	mkdir -p $(BUILD)
	python3 tools/prepare_media.py
	touch $@

$(BUILD)/boot.bin: boot/boot.asm
	mkdir -p $(BUILD)
	$(NASM) -f bin -o $@ $<

$(BUILD)/stage2.bin: boot/stage2.asm
	mkdir -p $(BUILD)
	$(NASM) -f bin -o $@ $<

$(BUILD)/start.o: kernel/start.asm
	mkdir -p $(BUILD)
	$(NASM) -f elf64 -o $@ $<

$(BUILD)/idt_asm.o: kernel/idt.asm
	$(NASM) -f elf64 -o $@ $<

$(BUILD)/frames.o: $(BUILD)/media.stamp
	cd $(BUILD) && $(OBJCOPY) -I binary -O elf64-x86-64 -B i386:x86-64 frames.rle frames.o

$(BUILD)/musicbin.o: $(BUILD)/media.stamp
	cd $(BUILD) && $(OBJCOPY) -I binary -O elf64-x86-64 -B i386:x86-64 music.bin musicbin.o

$(BUILD)/%.o: kernel/%.c
	mkdir -p $(BUILD)
	$(CC) $(CFLAGS) -c -o $@ $<

$(BUILD)/kernel.elf: linker.ld $(KERNEL_OBJS)
	$(LD) -nostdlib -static -z max-page-size=0x1000 -T linker.ld -o $@ $(KERNEL_OBJS)

$(BUILD)/kernel.bin: $(BUILD)/kernel.elf
	$(OBJCOPY) -O binary $< $@

$(BUILD)/chimia.img: $(BUILD)/boot.bin $(BUILD)/stage2.bin $(BUILD)/kernel.bin tools/mkimage.py
	python3 tools/mkimage.py

smoke: $(BUILD)/chimia.img
	python3 tools/smoke.py

run: $(BUILD)/chimia.img
	python3 tools/run.py

clean:
	rm -rf $(BUILD)
