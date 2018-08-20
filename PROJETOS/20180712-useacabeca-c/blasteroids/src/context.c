#include <blasteroids/config.h>
#include <blasteroids/bullet.h>
#include <blasteroids/asteroid.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/utils.h>
#include <blasteroids/collision.h>
#include <blasteroids/pos_fixer.h>
#include <blasteroids/text.h>
#include <blasteroids/config.h>

#include <blasteroids/context.h>
void blasteroids_context_tick(GameContext *ctx) {
    blasteroids_asteroid_update_all(ctx->asteroids->next);
    blasteroids_bullet_update_all(ctx->bullets, ctx->HearthBeat);
    blasteroids_context_update(ctx);
    blasteroids_context_draw(ctx);
}

void blasteroids_context_update(GameContext *ctx) {
    if (blasteroids_is_collision(ctx)) {
        blasteroids_asteroid_gc(ctx->asteroids);
        blasteroids_bullet_gc(ctx->bullets);
    }
    blasteroids_fix_positions(ctx);
}

void blasteroids_context_draw(GameContext *ctx) {
    al_flip_display();
    al_clear_to_color(al_map_rgb(0, 0, 0));
    blasteroids_ship_draw(ctx->ship);
    blasteroids_asteroid_draw_all(ctx->asteroids);
    blasteroids_bullet_draw_all(ctx->bullets);
    draw_life(ctx);
    draw_score(ctx);
    blasteroids_asteroid_draw_life(ctx);
#ifdef DEBUG_DRAW_COUNTER
    draw_counter(ctx);
#endif
}

int blasteroids_display_w(GameContext *ctx) {
    return al_get_display_width(ctx->display);
}

int blasteroids_display_h(GameContext *ctx) {
    return al_get_display_height(ctx->display);
}
