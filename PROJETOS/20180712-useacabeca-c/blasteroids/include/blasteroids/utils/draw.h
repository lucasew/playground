#ifndef _BLASTEROIDS_UTILS_DRAW
#define _BLASTEROIDS_UTILS_DRAW

#include <allegro5/allegro.h>

float deg2rad(float deg);

float blasteroids_get_delta_x(float speed, float degrees);

float blasteroids_get_delta_y(float speed, float degrees);

float get_distance(float ax, float ay, float bx, float by);

float al_draw_line_scaled(float ax, float ay, float bx, float by, ALLEGRO_COLOR color, float thickness, float scale);

#define al_draw_line_scaled(ax, ay, bx, by, color, thick, scale) \
    al_draw_line(ax*scale, ay*scale, bx*scale, by*scale, color, thick)

#endif
