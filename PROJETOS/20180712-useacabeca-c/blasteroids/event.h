#ifndef BLASTEROIDS_EVENT
#define BLASTEROIDS_EVENT

void event_loop_once(GameContext *ctx, ALLEGRO_EVENT *event);

void handle_event(ALLEGRO_EVENT *ev, GameContext *ctx);

#endif
