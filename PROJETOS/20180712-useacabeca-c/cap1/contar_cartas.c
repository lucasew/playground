#include<stdio.h>
#include<stdlib.h>

int get_ncarta(char carta[3]) {
    switch (carta[0]) {
        case 'K': return 13; break;
        case 'Q': return 12; break;
        case 'J': return 11; break;
        case 'A': return 1; break;
        default: return atoi(carta);
    };
};

int main() {
    char carta[3];
    printf("Digite uma carta: ");
    scanf("%2s", carta);
    int contador = 0;
    int ncarta = get_ncarta(carta);
    if (ncarta >= 3 && ncarta <= 6) { 
        printf("Incrementa \n");
    } else if (ncarta >= 10 && ncarta <= 13) {
        printf("Decrementa\n");
    } else {
        printf("Não faz nada\n");
    };
};


