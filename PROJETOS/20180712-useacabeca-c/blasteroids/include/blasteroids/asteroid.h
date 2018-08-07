#ifndef BLASTEROIDS_ASTEROID
#define BLASTEROIDS_ASTEROID

#include <blasteroids/blasteroids_types.h>

const float asteroid_points[ASTEROID_SEGMENTS][2];

void blasteroids_asteroid_draw(Asteroid *a);

void blasteroids_asteroid_draw_all(Asteroid *a);

void blasteroids_asteroid_update(Asteroid *a);

void blasteroids_asteroid_update_all(Asteroid *a);

void blasteroids_asteroid_append(Asteroid *old, Asteroid new); // Sem malloc no new

void blasteroids_destroy_asteroid(Asteroid *a);

void blasteroids_asteroid_gc(Asteroid *a);

#endif
