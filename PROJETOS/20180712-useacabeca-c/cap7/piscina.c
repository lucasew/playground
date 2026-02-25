#include <stdio.h>

enum response_type {
    DUMP,
    SECOND_CHANCE,
    MARRIAGE
};

typedef struct {
    char *name;
    enum response_type type;
} response;

void cumprimentar(response r) {
    printf("Dear %s,\n", r.name);
}

void dump(response r) {
    cumprimentar(r);
    puts("Unfortunately your last date contacted us to");
    puts("say that they will not be seeing you again."); // Força soldado kkj
}

void second_chance(response r) {
    cumprimentar(r);
    puts("Good news: your last date has asked us to");
    puts("arrange another meeting. Please call ASAP.");
}

void marriage(response r) {
    cumprimentar(r);
    puts("Congratulations! Your last date has contacted");
    puts("us with a proposal of marriage.");;
}

int main() {
    response r[] = {
        {"Mike", DUMP},
        {"Luis", SECOND_CHANCE},
        {"Matt", SECOND_CHANCE},
        {"William", MARRIAGE}
    };
    int i;
    for (i = 0; i < 4; i++) {
        switch(r[i].type) { // break tá quebrando o loop
            case DUMP:
                dump(r[i]);
            case SECOND_CHANCE: 
                second_chance(r[i]); 
            default: 
                marriage(r[i]);
        }
        return 0;
    }
}
