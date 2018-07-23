#include <stdio.h>

typedef struct island {
    char *name;
    char *opens;
    char *closes;
    struct island *next;
} island;

void island_display(island *start) {
    island *i = start;
    for (; i != NULL; i = i->next) {
        printf("Name: %s open: %s-%s\n", i->name, i->opens, i->closes);
    };
}

int main () {
    island amity = {"Amity", "09:00", "17:000", NULL};
    island craggy = {"Craggy", "09:00", "17:00", NULL};
    island isla_nublar = {"Isla Nublar", "09:00", "17:00",NULL};
    island shutter = {"Shutter", "09:00", "17:00", NULL};
    amity.next = &craggy;
    craggy.next = &isla_nublar;
    isla_nublar.next = &shutter;
    island_display(&amity);
}
