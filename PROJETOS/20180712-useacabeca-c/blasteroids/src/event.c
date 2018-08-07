#include <blasteroids/event.h>
#include <blasteroids.h>

void event_loop_once(GameContext *ctx, ALLEGRO_EVENT *event) {
    al_wait_for_event(ctx->event_queue, event);
    al_lock_mutex(ctx->mutex);
    handle_event(event, ctx);
    al_unlock_mutex(ctx->mutex);
}

void handle_event(ALLEGRO_EVENT *ev, GameContext *ctx) {
    if(ev->type == ALLEGRO_EVENT_KEY_DOWN) {
        switch (ev->keyboard.keycode) {
            case ALLEGRO_KEY_LEFT:
                blasteroids_ship_left(ctx->ship);
                return;
            case ALLEGRO_KEY_RIGHT:
                blasteroids_ship_right(ctx->ship);
                return;
            case ALLEGRO_KEY_UP:
                blasteroids_ship_up(ctx->ship);
                return;
            case ALLEGRO_KEY_DOWN:
                blasteroids_ship_down(ctx->ship);
                return;
            case ALLEGRO_KEY_ESCAPE:
                info("Sair");
                stop(0);
                return;
        }
    }
    if(ev->type == ALLEGRO_EVENT_TIMER) {
        ctx->HearthBeat = ctx->HearthBeat + 1;
        update_states(ctx);
        if (ctx->HearthBeat%10 == 0) blasteroids_asteroid_gc(ctx->asteroids);
    }
    if(ev->type == ALLEGRO_EVENT_DISPLAY_CLOSE) {
        stop(0);
        return;
    }
}
