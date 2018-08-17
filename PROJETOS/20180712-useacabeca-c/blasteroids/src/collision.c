#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>

#include <blasteroids/config.h>
#include <blasteroids/main.h>
#include <blasteroids/asteroid.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/bullet.h>
#include <blasteroids/utils/draw.h>
#include <blasteroids/context.h>
#include <blasteroids/utils.h>

#include <blasteroids/collision.h>

int blasteroids_check_collision_asteroid_spaceship(GameContext *ctx) {
    int collisions = 0;
    if (ctx->asteroids->next == NULL) return collisions;
    float sx, sy; // Centro da nave
    float ax, ay; // Centro de um asteroide
    sx = ctx->ship->sx;
    sy = ctx->ship->sy;
    float cur_distance, min_distance;
    Asteroid *this = ctx->asteroids->next;
    while (this != NULL) {
        ax = this->sx;
        ay = this->sy;
        cur_distance = get_distance(sx, sy, ax, ay);
        min_distance = 10 + 22*this->scale;
#ifdef DEBUG_COLLISION_GRAPH
        // Linha entre o asteroide e a nave
        ALLEGRO_TRANSFORM t;
        al_identity_transform(&t); // Base canônica
        al_use_transform(&t);
        al_draw_line(sx, sy, ax, ay, al_map_rgb(255, 255, 255), 1);
#endif
        if (!(cur_distance > min_distance)) {
            ctx->ship->health = ctx->ship->health - 1;
            this->health = this->health - 1;
            collisions++;
        }
        this = this->next;
    }
    return collisions;
}

int blasteroids_check_collision_asteroid_bullet(GameContext *ctx) {
    int collisions = 0;
    Asteroid *as = ctx->asteroids->next;
    Bullet *bu = ctx->bullets->next;
    if (as == NULL || bu == NULL) return collisions;
    float distancia;
    while (as != NULL) {
        bu = ctx->bullets;
        while (bu != NULL) {
            distancia = get_distance(as->sx, as->sy, bu->sx, bu->sy);
            if (distancia < (22*as->scale)) {
                as->health = as->health - bu->power;
                bu->health = 0;
                collisions++;
            }
            bu = bu->next;
        }
        as = as->next;
    }
    return collisions;

}

int blasteroids_is_collision(GameContext *ctx) {
#ifndef ASTEROID_SEGMENTS
    error("Constantes não definidas em teste de colisão");
#endif
    int collisions = 0;
    collisions += blasteroids_check_collision_asteroid_bullet(ctx);
    collisions += blasteroids_check_collision_asteroid_spaceship(ctx);
    debug("Collisions %i", collisions);
    return collisions;
}
