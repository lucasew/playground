#include <blasteroids.h>

void draw_life(GameContext *ctx) {
    ALLEGRO_TRANSFORM t;
    al_identity_transform(&t);
    al_use_transform(&t);
    al_draw_textf(ctx->font, al_map_rgb(255, 0, 0), 10, 10, ALLEGRO_ALIGN_LEFT, "<3 %i", ctx->ship->health);
}
