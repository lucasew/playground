#include <math.h>
#include <blasteroids/config.h>
#include <blasteroids/utils/draw.h>

float deg2rad(float deg) {
    return 0.0174532925 * deg;
}

float blasteroids_get_delta_x(float speed, float degrees) {
    return speed * sin((double)deg2rad(degrees));
}

float blasteroids_get_delta_y(float speed, float degrees) {
    return speed * cos((double)deg2rad(degrees)) * -1;
}

float get_distance(float ax, float ay, float bx, float by) {
    float x, y;
    x = ax - bx; // Não temos distancia negativa
    y = ay - by;
    float dist = sqrtf(x*x + y*y); // Vi va pitágoras :p
}

