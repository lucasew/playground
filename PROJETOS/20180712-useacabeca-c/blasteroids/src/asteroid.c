#include <stdlib.h>
#include <stdio.h>
#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>

#include <blasteroids/main.h>
#include <blasteroids/asteroid.h>
#include <blasteroids/utils.h>

void _log_asteroid(char *reason, struct Asteroid *a) {
#ifdef DEBUG_ASTEROID
    debug("asteroid %s (%f, %f) heading:%f speed:%f rot_velocity:%f scale:%f health:%i", reason, a->sx, a->sy, a->heading, a->speed, a->rot_velocity, a->scale, a->health);
#endif
}

const float asteroid_points[ASTEROID_SEGMENTS][2] = {
    {-20, 20},
    {-25, 5},
    {-25, -10},
    {-5, -10},
    {-10, -20},
    {5, -20},
    {20, -10},
    {20, -5},
    {0, 0},
    {20, 10},
    {10, 20},
    {0, 15}
};

void blasteroids_asteroid_draw(struct Asteroid *a) {
    if (a->sx > 0 && a->sy > 0) {
        ALLEGRO_TRANSFORM transform;
        al_identity_transform(&transform);
        al_rotate_transform(&transform, deg2rad(a->heading));
        al_translate_transform(&transform, a->sx, a->sy);
        al_use_transform(&transform);
        int i;
        for (i = 0; i < (ASTEROID_SEGMENTS); i++) {
            al_draw_line_scaled(
                    asteroid_points[i][0],
                    asteroid_points[i][1],
                    asteroid_points[(i + 1)%ASTEROID_SEGMENTS][0], // O módulo é para quando ele chegar no final da lista
                    asteroid_points[(i + 1)%ASTEROID_SEGMENTS][1],
                    a->color, 2.0f, a->scale);
        }
    }
}

void blasteroids_asteroid_draw_all(struct Asteroid *a) {
    struct Asteroid *tmp = a;
    while (tmp != NULL) {
        blasteroids_asteroid_draw(tmp);
        tmp = tmp->next;
    }
}

void blasteroids_asteroid_draw_life(GameContext *ctx) {
    struct Asteroid *a = ctx->asteroids->next; // O primeiro só tá lá pra facilitar
    while (a != NULL) {
        ALLEGRO_TRANSFORM t;
        al_identity_transform(&t);
        al_translate_transform(&t, a->sx, a->sy);
        al_use_transform(&t);
        al_draw_textf(ctx->font, al_map_rgb(255, 0, 0), 0, 0, ALLEGRO_ALIGN_CENTER, "%i", a->health);
        a = a->next;
    }
}

void blasteroids_asteroid_update(struct Asteroid *a) {
    float deltax, deltay;
    _log_asteroid("before", a);
    a->heading = a->heading + a->rot_velocity;
    a->sx = a->sx + blasteroids_get_delta_x(a->speed, a->heading);
    a->sy = a->sy + blasteroids_get_delta_y(a->speed, a->heading);
    _log_asteroid("after", a);
}

void blasteroids_asteroid_update_all(struct Asteroid *a) {
    struct Asteroid *this = a;
    while (this != NULL) {
        blasteroids_asteroid_update(this);
        this = this->next;
    }
}

void blasteroids_asteroid_append(struct Asteroid *old, struct Asteroid new) {//  Não é necessário dar malloc
    struct Asteroid *tmp = malloc(sizeof(struct Asteroid));
    *tmp = new;
    if (old->next != NULL) {
        tmp->next = old->next;
    }
    old->next = tmp;
}

void blasteroids_destroy_asteroid(struct Asteroid *a) {
    struct Asteroid *dummy;
    while (a != NULL) {
        dummy = a;
        a = a->next;
        free(dummy);
    }
}

void blasteroids_asteroid_gc(struct Asteroid *a) {
    debug("Removendo asteroides destruidos da memória...");
    struct Asteroid *previous = a;
    a = a->next;
    while (a != NULL) {
        if (a->health <= 0) {
            previous->next = a->next;
            free(a);
        }
        previous = a;
        a = a->next;
    }
}

