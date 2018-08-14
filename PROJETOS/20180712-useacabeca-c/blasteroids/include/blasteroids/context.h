#ifndef _BLASTEROIDS_CONTEXT
#define _BLASTEROIDS_CONTEXT

#include <allegro5/allegro.h>
#include <allegro5/allegro_font.h>

struct GameContext {
    ALLEGRO_DISPLAY *display;
    ALLEGRO_EVENT_QUEUE *event_queue;
    ALLEGRO_TIMER *timer;
    ALLEGRO_MUTEX *mutex;
    ALLEGRO_FONT *font;
    struct Spaceship *ship;
    struct Asteroid *asteroids;
    short lifes;
    int HearthBeat;
};

typedef struct GameContext GameContext;

#endif
