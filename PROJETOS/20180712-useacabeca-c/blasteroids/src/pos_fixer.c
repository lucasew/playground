#include <blasteroids/pos_fixer.h>
#include <blasteroids/context.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/asteroid.h>
#include <blasteroids/bullet.h>

void blasteroids_fix_positions(GameContext *ctx) {
    int h = blasteroids_display_h(ctx);
    int w = blasteroids_display_w(ctx);
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
