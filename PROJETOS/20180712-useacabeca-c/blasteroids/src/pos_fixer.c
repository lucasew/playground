#include <blasteroids/pos_fixer.h>
#include <blasteroids/context.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/asteroid.h>
#include <blasteroids/bullet.h>

void blasteroids_fix_positions(GameContext *ctx) {
    int h = al_get_display_height(ctx->display);
    int w = al_get_display_width(ctx->display);
    INSIDE_SCREEN(ctx->ship, w, h);
    Asteroid *adummy = ctx->asteroids->next;
    while (adummy != NULL) {
        INSIDE_SCREEN(adummy, w, h);
        adummy = adummy->next;
    }
    Bullet *bdummy = ctx->bullets->next;
    while (bdummy != NULL) {
        INSIDE_SCREEN(bdummy, w, h);
        bdummy = bdummy->next;
    }
}
