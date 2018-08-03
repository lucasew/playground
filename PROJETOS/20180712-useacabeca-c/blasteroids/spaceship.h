#ifndef BLASTEROIDS_SPACESHIP
#define BLASTEROIDS_SPACESHIP

// Usaremos isso para as caixas de colisão
#define SPACESHIP_SIZE_X 16
#define SPACESHIP_SIZE_Y 20

typedef struct Spaceship {
    float sx;
    float sy;
    float heading;
    float speed;
    bool gone;
    ALLEGRO_COLOR color;
} Spaceship;

void blasteroids_ship_draw(struct Spaceship *s);

void blasteroids_ship_left(Spaceship *s);

void blasteroids_ship_right(Spaceship *s);

void blasteroids_ship_down(Spaceship *s);

void blasteroids_ship_up(Spaceship *s);

#endif
