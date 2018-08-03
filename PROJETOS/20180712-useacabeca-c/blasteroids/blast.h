#ifndef BLASTEROIDS_BLAST
#define BLASTEROIDS_BLAST

typedef struct Blast {
    float sx;
    float sy;
    float heading;
    float speed;
    bool isOn;
    ALLEGRO_COLOR color;
} Blast;

#endif
