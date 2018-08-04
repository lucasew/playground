#ifndef BLASTEROIDS_ASTEROID
#define BLASTEROIDS_ASTEROID

#include <blasteroids/blasteroids_types.h>

void blasteroids_asteroid_draw(Asteroid *a);

void blasteroids_AsteroidNode_draw_all(AsteroidNode *node);

void blasteroids_asteroid_update(Asteroid *a);

void blasteroids_AsteroidNode_update_all(AsteroidNode *node);

void blasteroids_destroy_AsteroidNode(AsteroidNode *node);

void blasteroids_AsteroidNode_gc(AsteroidNode *node);

void blasteroids_asteroid_get_center(float *cx, float *cy, Asteroid *a);

#endif
