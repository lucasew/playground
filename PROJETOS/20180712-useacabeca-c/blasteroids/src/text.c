#include <allegro5/allegro.h>
#include <allegro5/allegro_ttf.h>
#include <blasteroids/context.h>
#include <blasteroids/text.h>
#include <blasteroids/spaceship.h>

void draw_life(GameContext *ctx) {
    ALLEGRO_TRANSFORM t;
    al_identity_transform(&t);
    al_use_transform(&t);
    al_draw_textf(ctx->font, al_map_rgb(0, 0, 255), 10, 10, ALLEGRO_ALIGN_LEFT, "<3 %i", ctx->ship->health);
}

void draw_counter(GameContext *ctx) {
    ALLEGRO_TRANSFORM t;
    al_identity_transform(&t);
    al_use_transform(&t);
    al_draw_textf(ctx->font, al_map_rgb(255, 255, 255), 300, 10, ALLEGRO_ALIGN_RIGHT, "C: %i", ctx->HearthBeat);
}

void draw_score(GameContext *ctx) {
    ALLEGRO_TRANSFORM t;
    al_identity_transform(&t);
    al_use_transform(&t);
    al_draw_textf(ctx->font, al_map_rgb(0, 255, 0), 500, 10, ALLEGRO_ALIGN_RIGHT, "PTS: %i", ctx->score);
}
