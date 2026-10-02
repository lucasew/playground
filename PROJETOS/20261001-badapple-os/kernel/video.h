#pragma once

#include <stdint.h>

void video_init(void);
int video_open(const uint8_t *blob);
void video_blit(uint32_t index);
uint32_t video_frame_count(void);
uint16_t video_fps(void);
void video_clear(void);
