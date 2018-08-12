#ifndef _BLASTEROIDS_POS_FIXER
#define _BLASTEROIDS_POS_FIXER

#include <blasteroids.h>

// Deixa os elementos dentro da janela
#define INSIDE_SCREEN(obj, w, h) \
    do { \
        if(obj->sx > w) { \
        obj->sx = 0; \
        }; \
        if (obj->sy > h) { \
            obj->sy = 0; \
        }; \
        if(obj->sx < 0) { \
            obj->sx = w; \
        }; \
        if(obj->sy < 0) {  \
            obj->sy = h; \
        }; \
    } while (0);

void blasteroids_fix_positions(GameContext *ctx);

#endif
