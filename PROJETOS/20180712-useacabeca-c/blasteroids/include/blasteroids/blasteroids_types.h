#ifndef BLASTEROIDS_TYPES
#define BLASTEROIDS_TYPES

#include <allegro5/allegro.h>

// =========== ASTEROID ============

// Caixas de colisão
// NÃO ESQUECE DE MULTIPLICAR PELO SCALE
#define ASTEROID_SIZE_X 35 // -20
#define ASTEROID_SIZE_Y 26 // -10

typedef struct Asteroid {
    float sx;
    float sy;
    float heading; // direção
    float speed; // velocidade
    float rot_velocity; // rotação por frame
    float scale; //tamanho * constante
    bool gone; // Destruído?
    ALLEGRO_COLOR color;
} Asteroid;

// =========== SPACESHIP ===========

#define SPACESHIP_SIZE_X 8 // -8
#define SPACESHIP_SIZE_Y 10 // -10
typedef struct Spaceship {
    float sx;
    float sy;
    float heading;
    float speed;
    bool gone;
    ALLEGRO_COLOR color;
} Spaceship;

typedef struct AsteroidNode {
    Asteroid *this;
    struct AsteroidNode *next;
} AsteroidNode;

// ========== BLAST ============

typedef struct Blast {
    float sx;
    float sy;
    float heading;
    float speed;
    bool isOn;
    ALLEGRO_COLOR color;
} Blast;

// ========= MAIN =============

typedef struct GameContext {
    ALLEGRO_DISPLAY *display;
    ALLEGRO_EVENT_QUEUE *event_queue;
    ALLEGRO_TIMER *timer;
    ALLEGRO_MUTEX *mutex;
    Spaceship *ship;
    AsteroidNode *asteroids;
    short lifes;
    int HearthBeat;
} GameContext;

#endif
