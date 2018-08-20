#include <stdlib.h>
#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>

#include <blasteroids/context.h>
#include <blasteroids/config.h>
#include <blasteroids/utils.h>
#include <blasteroids/spaceship.h>
#include <blasteroids/bullet.h>

void _log_bullet(char *reason, struct Bullet *b) {
    debug("bullet %s (%f, %f) h:%f s:%f pw:%i", reason, b->sx, b->sy, b->heading, b->speed, b->power);
}

void blasteroids_bullet_draw(struct Bullet *b) {
    ALLEGRO_TRANSFORM t;
    al_identity_transform(&t);
    al_rotate_transform(&t, deg2rad(b->heading));
    al_translate_transform(&t, b->sx, b->sy);
    al_use_transform(&t);
    al_draw_line(1, 0, 0, 1, b->color, 2.0f);
    al_draw_line(0, 1, -1, 0, b->color, 2.0f);
    al_draw_line(-1, 0, 0, -1, b->color, 2.0f);
    al_draw_line(0, -1, 1, 0, b->color, 2.0f);
}

void blasteroids_bullet_draw_all(struct Bullet *b) {
    struct Bullet *tmp = b->next;
    while (tmp != NULL) {
        blasteroids_bullet_draw(tmp);
        tmp = tmp->next;
    }
}

void blasteroids_bullet_update(struct Bullet *b, int HearthBeat) {
    b->sx = b->sx + blasteroids_get_delta_x(b->speed, b->heading)/FPS;
    b->sy = b->sy + blasteroids_get_delta_y(b->speed, b->heading)/FPS;
    if(!(HearthBeat%60)) {
        b->power--;
        _log_bullet("update", b);
    }
}

void blasteroids_bullet_update_all(struct Bullet *b, int HearthBeat) {
    struct Bullet *this = b->next; // Next ignora aquele bullet genesis
    while (this != NULL) {
        blasteroids_bullet_update(this, HearthBeat);
        this = this->next;
    }
}

void blasteroids_bullet_append(struct Bullet *old, struct Bullet new) { // Eu dou malloc sozinho
    _log_bullet("append", &new);
    struct Bullet *tmp = malloc(sizeof(struct Bullet));
    *tmp = new;
    if (old->next != NULL) {
        tmp->next = old->next;
    }
    old->next = tmp;
}

void blasteroids_bullet_shot(struct GameContext *ctx) {
    Bullet bt;
    bt.sx = ctx->ship->sx;
    bt.sy = ctx->ship->sy;
    bt.heading = ctx->ship->heading;
    bt.speed = 1 + rand()%100;
    bt.power = 1 + rand()%50;
    bt.color = al_map_rgb(rand()%255, rand()%255, rand()%255);
    bt.next = NULL;
    blasteroids_bullet_append(ctx->bullets, bt);   
}

void blasteroids_destroy_bullet(struct Bullet *b) {
    struct Bullet *dummy;
    while (b != NULL) {
        dummy = b;
        b = b->next;
        free(dummy);
    }
}

void blasteroids_bullet_gc(struct Bullet *b) {
    debug("Removendo tiros destruidos da memória...");
    if (b == NULL) return;
    struct Bullet *dummy, *previous = b, *this = b->next;
    while (this != NULL) {
        if (this->power <= 0) {
            dummy = this;
            previous->next = this->next;
            free(dummy);
            return;
        }
        previous = this;
        this = this->next;
    }
}
