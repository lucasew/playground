#ifndef BLASTEROIDS_ASTEROID
#define BLASTEROIDS_ASTEROID

// Usaremos isso para as caixas de colisão
#define ASTEROID_SIZE_X 55
#define ASTEROID_SIZE_Y 40

typedef struct Asteroid {
    float sx;
    float sy;
    float heading; // direção
    float speed; // velocidade
    float rot_velocity; // rotação por frame;
    float scale; // tamanho*c
    bool gone; // Destruído?
    ALLEGRO_COLOR color;
} Asteroid;

typedef struct AsteroidNode {
    Asteroid *this;
    struct AsteroidNode *next;
} AsteroidNode;

void blasteroids_asteroid_draw(Asteroid *a);

void blasteroids_AsteroidNode_draw_all(AsteroidNode *node);

void blasteroids_asteroid_update(Asteroid *a);

void blasteroids_AsteroidNode_update_all(AsteroidNode *node);

void blasteroids_destroy_AsteroidNode(AsteroidNode *node);

void blasteroids_AsteroidNode_gc(AsteroidNode *node);

#endif
