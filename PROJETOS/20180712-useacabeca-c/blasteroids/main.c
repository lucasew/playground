#define DISPLAY_ALTURA 600
#define DISPLAY_LARGURA 600

#include <blasteroids/main.h>
#include <blasteroids/utils.h>
#include <blasteroids.h>
#include <signal.h>
#include <stdio.h>

const char *WindowTitle = "BLASTEROIDS by Lucas59356";
bool *running;
GameContext *ctx;

bool is_collision(GameContext *ctx) {
    float sx, sy; // Centro da nave
    blasteroids_spaceship_get_center(&sx, &sy, ctx->ship);
    AsteroidNode *this;
    this = ctx->asteroids;
    float ax, ay;
    while (this != NULL) {
        blasteroids_asteroid_get_center(&ax, &ay, this->this); // Tem que estar perto suficiente para dar dano
        if (get_distance(sx, sy, ax, ay) < ((ASTEROID_SIZE_X + ASTEROID_SIZE_Y)/2 + (SPACESHIP_SIZE_X + SPACESHIP_SIZE_Y)/2))
            return true;
        this = this->next;
    }
    return false;
}

void update_states() {
    blasteroids_AsteroidNode_update_all(ctx->asteroids);
}

int main() {
    info("Iniciando...");
    running = malloc(sizeof(bool)); // Se isso for falso os loops param e o programa começa a desligar
    *running = true;
    // Signal handler
    if(catch_signal(SIGINT, stop) == -1 || catch_signal(SIGTERM, stop) == -1)
        error("Não foi possível setar o handler de interrupção");
    // Criando o contexto
    ctx = malloc(sizeof(GameContext));
    info("Inicializando...");
    if (!al_init())
        error("Não foi possível inicializar biblioteca de suporte!");
    if(!al_init_primitives_addon())
        error("Não foi possível inicializar a primitives addon");
    // Queue
    ctx->event_queue = al_create_event_queue();
    // Timer
    ctx->timer = al_create_timer(1);
    if (!ctx->timer)
        error("Não foi possível iniciar o timer");
    al_start_timer(ctx->timer);
    al_register_event_source(ctx->event_queue, al_get_timer_event_source(ctx->timer));
    // Mutex
    ctx->mutex = al_create_mutex();
    if (!ctx->mutex)
        error("Não foi possível criar o mutex");
    // Teclado
    if(!al_install_keyboard())
        error("Não foi possível conectar-se ao teclado");
    al_register_event_source(ctx->event_queue, al_get_keyboard_event_source());
    // Display
    //al_set_new_display_flags(ALLEGRO_FULLSCREEN_WINDOW);
    ctx->display = al_create_display(DISPLAY_ALTURA, DISPLAY_LARGURA);
    al_set_window_title(ctx->display, WindowTitle); // Título da janela
    al_register_event_source(ctx->event_queue, al_get_display_event_source(ctx->display));
    // Criando spaceship de exemplo
    Spaceship *sp = malloc(sizeof(Spaceship));
    sp->sx = 200;
    sp->sy = 200;
    sp->heading = 20;
    sp->speed = 10;
    sp->gone = false;
    sp->color = al_map_rgb(255, 255, 0);
    ctx->ship = sp;
    // Criando asteroide de exemplo
    ctx->asteroids = malloc(sizeof(AsteroidNode));
    Asteroid *as = malloc(sizeof(Asteroid));
    ctx->asteroids->this = as;
    ctx->asteroids->next = NULL;
    as->sx = 300.0;
    as->sy = 350.0;
    as->heading = 230.0;
    as->speed = 12.0;
    as->rot_velocity = 5.0;
    as->scale = 1.5;
    as->gone = false;
    as->color = al_map_rgb(15, 135, 88);
    // Event loop in main thread
    ALLEGRO_EVENT event;
    while(1) {
        if (!*running) break;
        al_flip_display();
        al_clear_to_color(al_map_rgb(0, 0, 0));
        blasteroids_ship_draw(ctx->ship);
        blasteroids_AsteroidNode_draw_all(ctx->asteroids);
        event_loop_once(ctx, &event);
    }
    // ============= SAINDO ===========
    handle_shutdown(SIGINT);
}
void handle_shutdown() {
    info("Saindo....");
    /*debug("Destroy timer");
      al_destroy_timer(ctx->timer);*/
    debug("Destroy mutex");
    al_destroy_mutex(ctx->mutex);
    debug("Destroy timer");
    al_destroy_timer(ctx->timer);
    // Queue
    debug("Destroy queue");
    al_destroy_event_queue(ctx->event_queue);
    debug("Free ship");
    free(ctx->ship);
    debug("Free running");
    free(running);
    debug("Free node");
    blasteroids_destroy_AsteroidNode(ctx->asteroids);
    debug("Destroy display");
    al_destroy_display(ctx->display);
    debug("Free ctx");
    free(ctx);
    //raise(SIGKILL);
    exit(1);
}

void stop(int sig) {
    *running = false;
}
