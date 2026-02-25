#include<stdio.h>

enum {
    PRIMEIRO = 1 << 0,
    SEGUNDO = 1 << 1,
    TERCEIRO = 1 << 2,
    QUARTO,
    QUINTO
};

int main() {
    printf("Numeros: %i %i %i %i %i\n", PRIMEIRO, SEGUNDO, TERCEIRO, QUARTO, QUINTO); // Não, ele não vai fazer aquilo que eu queria que ele fizesse :(
    return 0;
}
