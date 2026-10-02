ENTRY(_start)

SECTIONS {
    . = 0x100000;

    .text : {
        KEEP(*(.text.boot))
        *(.text)
        *(.text.*)
    }

    .rodata : ALIGN(16) {
        *(.rodata)
        *(.rodata.*)
    }

    .data : ALIGN(16) {
        *(.data)
        *(.data.*)
    }

    .bss : ALIGN(16) {
        bss_start = .;
        *(.bss)
        *(.bss.*)
        *(COMMON)
        bss_end = .;
    }

    /DISCARD/ : {
        *(.comment)
        *(.eh_frame)
        *(.eh_frame_hdr)
        *(.note*)
        *(.gnu*)
    }
}
