#ifndef _BLASTEROIDS_SPAWNER
#define _BLASTEROIDS_SPAWNER

#include <blasteroids/config.h>

#define RAND_COLOR rand()%200 + 55

void blasteroids_asteroid_generate(GameContext *ctx);

#endif
