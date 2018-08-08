#include <stdlib.h>
#include <stdio.h>
#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>

#include <blasteroids/blasteroids_types.h>
#include <blasteroids.h>
#include <blasteroids/asteroid.h>

void _log_asteroid(char *reason, Asteroid *a) {
    debug("asteroid %s (%f, %f) heading:%f speed:%f rot_velocity:%f scale:%f health:%i", reason, a->sx, a->sy, a->heading, a->speed, a->rot_velocity, a->scale, a->health);
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

void blasteroids_asteroid_draw(Asteroid *a) {
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

void blasteroids_asteroid_draw_all(Asteroid *a) {
    Asteroid *tmp = a;
    for (;;) {
        if (tmp == NULL) return;
        blasteroids_asteroid_draw(tmp);
        tmp = tmp->next;
    }
}

void blasteroids_asteroid_update(Asteroid *a) {
    float deltax, deltay;
    _log_asteroid("before", a);
    a->heading = a->heading + a->rot_velocity;
    a->sx = a->sx + blasteroids_get_delta_x(a->speed, a->heading);
    a->sy = a->sy + blasteroids_get_delta_y(a->speed, a->heading);
    _log_asteroid("after", a);
}

void blasteroids_asteroid_update_all(Asteroid *a) {
    Asteroid *this = a;
    for (;;) {
        if(this == NULL) return;
        blasteroids_asteroid_update(this);
        this = this->next;
    }
}

void blasteroids_asteroid_append(Asteroid *old, Asteroid new) {//  Não é necessário dar malloc
    Asteroid *tmp = malloc(sizeof(Asteroid));
    *tmp = new;
    if (old->next != NULL) {
        tmp->next = old->next;
    }
    old->next = tmp;
}

void blasteroids_destroy_asteroid(Asteroid *a) {
    for(;;) {
        if (a == NULL) return;
        blasteroids_destroy_asteroid(a->next);
        free(a);
        a = a->next;
    }
}

void blasteroids_asteroid_gc(Asteroid *a) {
    debug("Removendo asteroides destruidos da memória...");
    Asteroid *previous = a;
    a = a->next;
    for (;;) {
        if (a == NULL) break;
        if (a->health <= 0) {
            previous->next = a->next;
            free(a);
        }
        previous = a;
        a = a->next;
    }
}

