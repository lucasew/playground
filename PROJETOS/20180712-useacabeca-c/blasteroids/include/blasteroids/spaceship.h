#ifndef BLASTEROIDS_SPACESHIP
#define BLASTEROIDS_SPACESHIP

#include <blasteroids/blasteroids_types.h>

void blasteroids_ship_draw(Spaceship *s);

void blasteroids_ship_left(Spaceship *s);

void blasteroids_ship_right(Spaceship *s);

void blasteroids_ship_down(Spaceship *s);

void blasteroids_ship_up(Spaceship *s);

void blasteroids_spaceship_get_center(float *cx, float *cy, Spaceship *s);

#endif
