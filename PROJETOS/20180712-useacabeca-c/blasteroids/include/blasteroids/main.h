#ifndef _BLASTEROIDS_MAIN
#define _BLASTEROIDS_MAIN
#include <stdbool.h>
#include <blasteroids/blasteroids_types.h>

int is_collision(GameContext *ctx);

void update_states();

void handle_shutdown();

void stop(int sig);

#endif
