#ifndef BLASTEROIDS_UTILS_DRAW
#define BLASTEROIDS_UTILS_DRAW

#define al_draw_line_scaled(ax, ay, bx, by, color, thick, scale) \
    al_draw_line((float)ax*scale, (float)ay*scale, (float)bx*scale, (float)by*scale, color, thick)

float deg2rad(float deg);

void blasteroids_get_delta(float *deltax, float *deltay, float speed, float degrees);

float get_distance(float ax, float ay, float bx, float by);

#endif
