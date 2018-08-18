#include <blasteroids/context.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/asteroid.h>
#include <blasteroids/bullet.h>
#include <stdlib.h>
#include <blasteroids/utils.h>
#include <blasteroids/main.h> // Função stop
#include <blasteroids/event.h>

void event_loop_once(GameContext *ctx, ALLEGRO_EVENT *event) {
    al_wait_for_event(ctx->event_queue, event);
    al_lock_mutex(ctx->mutex);
    handle_event(event, ctx);
    al_unlock_mutex(ctx->mutex);
}

void handle_event(ALLEGRO_EVENT *ev, GameContext *ctx) {
    Bullet bt;
    if(ev->type == ALLEGRO_EVENT_KEY_DOWN) {
        switch (ev->keyboard.keycode) {
            case ALLEGRO_KEY_LEFT:
                blasteroids_ship_left(ctx->ship);
                blasteroids_context_update(ctx);
                return;
            case ALLEGRO_KEY_RIGHT:
                blasteroids_ship_right(ctx->ship);
                blasteroids_context_update(ctx);
                return;
            case ALLEGRO_KEY_UP:
                blasteroids_ship_up(ctx->ship);
                blasteroids_context_update(ctx);
                return;
            case ALLEGRO_KEY_DOWN:
                blasteroids_ship_down(ctx->ship);
                blasteroids_context_update(ctx);
                return;
            case ALLEGRO_KEY_ESCAPE:
                info("Sair");
                stop(0);
                return;
            case ALLEGRO_KEY_SPACE:
                /*if (ctx->HearthBeat%2)*/ blasteroids_bullet_shot(ctx);
                return;
        }
    }
    if(ev->type == ALLEGRO_EVENT_TIMER) {
        ctx->HearthBeat = ctx->HearthBeat + 1;
        blasteroids_context_tick(ctx);
        if (!(ctx->HearthBeat%10)) blasteroids_asteroid_generate(ctx);
    }
    if(ev->type == ALLEGRO_EVENT_DISPLAY_CLOSE) {
        stop(0);
        return;
    }
    if(ev->type == ALLEGRO_EVENT_DISPLAY_RESIZE) {
        if (!al_acknowledge_resize(ctx->display))
            error("Não foi possível redimensionar o display");
    }
}
