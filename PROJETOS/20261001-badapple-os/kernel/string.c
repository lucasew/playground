#include <stddef.h>
#include <stdint.h>

void *memset(void *dst, int value, size_t n) {
    unsigned char *bytes = dst;
    unsigned char byte = (unsigned char)value;
    /* Byte stores to the VGA window are slow on real chipsets. Fill the
     * aligned bulk with one stosd so a frame does not stall the song. */
    while (n && ((uintptr_t)bytes & 3u)) {
        *bytes++ = byte;
        n--;
    }
    if (n >= 4) {
        uint32_t chunk = 0x01010101u * byte;
        size_t dwords = n / 4;
        uint32_t *wide = (uint32_t *)bytes;
        __asm__ volatile("rep stosl"
                         : "+D"(wide), "+c"(dwords)
                         : "a"(chunk)
                         : "memory");
        bytes = (unsigned char *)wide;
        n &= 3u;
    }
    while (n--) {
        *bytes++ = byte;
    }
    return dst;
}

void *memcpy(void *dst, const void *src, size_t n) {
    unsigned char *d = dst;
    const unsigned char *s = src;
    while (n--) {
        *d++ = *s++;
    }
    return dst;
}
