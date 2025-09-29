#include<stdio.h>

int main () {
    float rendimento, gasto;
    int distancia;
    scanf("%i %f", &distancia, &gasto);
    rendimento = (float)distancia / gasto;
    printf("%.3f km/l\n", rendimento);
    return 0;
}
