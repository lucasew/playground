#include<stdio.h>

void go_south_east(int *lat, int *lon) {
    *lat = *lat - 1;
    *lon = *lon + 1;
}

int main() {
    int lat = 32;
    int lon = -64;
    go_south_east(&lat, &lon);
    printf("Estamos agora em (%i, %i)!\n", lat, lon);
    return 0;
}
