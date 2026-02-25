#include <stdio.h>

typedef enum {
    PRIMEIRO,
    SEGUNDO,
    TERCEIRO = 12,
    QUARTO
} valores;

int main() {
    printf("%i\n", PRIMEIRO);
    printf("%i\n", SEGUNDO);
    printf("%i\n", TERCEIRO);
    printf("%i\n", QUARTO);
}
