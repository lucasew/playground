#include "log_utils.h"
#include "draw_utils.h"
#include <math.h>

float deg2rad(float deg) {
    return 0.0174532925 * deg;
}

void blasteroids_get_delta(float *deltax, float *deltay, float speed, float degrees) {
    *deltax = speed * sin((double)deg2rad(degrees));
    *deltay = speed * cos((double)deg2rad(degrees)) * -1;
}
