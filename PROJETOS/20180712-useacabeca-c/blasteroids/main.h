#ifndef BLASTEROIDS_MAIN
#define BLASTEROIDS_MAIN

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

void update_states();

void handle_shutdown();

void stop(int sig);

#endif
