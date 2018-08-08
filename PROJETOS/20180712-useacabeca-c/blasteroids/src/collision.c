#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>
#include <blasteroids/config.h>
#include <blasteroids/blasteroids_types.h>
#include <blasteroids/utils/draw.h>
#include <blasteroids/utils/log.h>

int is_collision(GameContext *ctx) {
#ifndef ASTEROID_SEGMENTS
    error("Constantes não definidas em teste de colisão");
#endif
    float sx, sy; // Centro da nave
    float ax, ay; // Centro do asteroide
    sx = ctx->ship->sx;
    sy = ctx->ship->sy;
    float cur_distance, min_distance;
    Asteroid *this = ctx->asteroids->next;
    for(;;) {
        if(this == NULL) break; // Para o continue não criar um loop infinito
        assert(this != NULL);
        ax = this->sx;
        ay = this->sy;
        cur_distance = get_distance(sx, sy, ax, ay);
        min_distance = 10 + 22*this->scale;
#ifdef DEBUG
        // debug, apenas uma linha que é traçada entre o asteroide e a nave
        ALLEGRO_TRANSFORM t; // Vamos converter para a base canônica
        al_identity_transform(&t);
        al_use_transform(&t);
        al_draw_line(sx, sy, ax, ay, al_map_rgb(255, 255, 255), 1); // Linha da distancia
        debug("collision: cur_distance:%f min_distance:%f", cur_distance, min_distance);
#endif
        if (!(cur_distance > min_distance)) {
            ctx->ship->health = ctx->ship->health - 1;
            this->health = this->health - 1;
            return 1;
        }
        this = this->next;
    }
    return 0;
}
