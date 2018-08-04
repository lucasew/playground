#include <math.h>
#include <blasteroids/utils/draw.h>

float deg2rad(float deg) {
    return 0.0174532925 * deg;
}

void blasteroids_get_delta(float *deltax, float *deltay, float speed, float degrees) {
    *deltax = speed * sin((double)deg2rad(degrees));
    *deltay = speed * cos((double)deg2rad(degrees)) * -1;
}


float get_distance(float ax, float ay, float bx, float by) {
    float x, y;
    x = ax - bx; // Não temos distancia negativa
    y = ay - by;
    return sqrtf(x*x + y*y); // Vi va pitágoras :p
}
