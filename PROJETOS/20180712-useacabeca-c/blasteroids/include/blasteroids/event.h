#ifndef BLASTEROIDS_EVENT
#define BLASTEROIDS_EVENT

#include <blasteroids/blasteroids_types.h>

void event_loop_once(GameContext *ctx, ALLEGRO_EVENT *event);

void handle_event(ALLEGRO_EVENT *ev, GameContext *ctx);

#endif
