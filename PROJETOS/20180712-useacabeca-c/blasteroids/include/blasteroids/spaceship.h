#ifndef _BLASTEROIDS_SPACESHIP
#define _BLASTEROIDS_SPACESHIP

// Quantos graus a nave vai virar a cada vez que apertamos direita ou esquerda
#define HEADING_STEP 10

// Colisão
#define SPACESHIP_SIZE_X 8 // -8
#define SPACESHIP_SIZE_Y 10 // -10

// Estrutura
struct Spaceship {
    float sx;
    float sy;
    float heading;
    float speed;
    int health;
    ALLEGRO_COLOR color;
};

typedef struct Spaceship Spaceship;

void blasteroids_ship_draw(Spaceship *s);

void blasteroids_ship_left(Spaceship *s);

void blasteroids_ship_right(Spaceship *s);

void blasteroids_ship_down(Spaceship *s);

void blasteroids_ship_up(Spaceship *s);

#endif
