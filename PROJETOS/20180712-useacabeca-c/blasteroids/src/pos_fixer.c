#include <blasteroids/pos_fixer.h>
#include <blasteroids/context.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/asteroid.h>

void blasteroids_fix_positions(GameContext *ctx) {
    int h = al_get_display_height(ctx->display);
    int w = al_get_display_width(ctx->display);
    INSIDE_SCREEN(ctx->ship, w, h);
    Asteroid *dummy = ctx->asteroids->next;
    for(;;) {
        if (dummy == NULL) break;
        INSIDE_SCREEN(dummy, w, h);
        dummy = dummy->next;
    }
}
