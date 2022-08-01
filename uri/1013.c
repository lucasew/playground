#include <stdio.h>

int abs(int n) {
    if (n < 0) {
        return n*-1;
    }
    return n;
}

int get_maior(int x, int y) {
    return (x + y + abs(x - y)) / 2;
}

int main() {
    int A, B, C;
    scanf("%i %i %i", &A, &B, &C);
    printf("%i eh o maior\n", get_maior(get_maior(A, B), C));
}
