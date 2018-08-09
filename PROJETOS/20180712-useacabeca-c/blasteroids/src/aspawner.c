#include <stdlib.h>
#include <blasteroids.h>
#include <blasteroids/aspawner.h>
#include <time.h>

void blasteroids_asteroid_generate(GameContext *ctx) {
    srand(time(NULL));
    Asteroid as;
    as.sx = rand() % DISPLAY_LARGURA;
    as.sx = rand() % DISPLAY_ALTURA;
    as.heading = rand() % 360;
    as.speed = (float)((rand() % 200)/10.0);
    as.rot_velocity = (float)(rand()%20);
    as.scale = (float)((rand()%40)/10) + 0.5;
    as.health = rand() % 200;
    as.color = al_map_rgb(RAND_COLOR, RAND_COLOR, RAND_COLOR);
    blasteroids_asteroid_append(ctx->asteroids, as);
}
