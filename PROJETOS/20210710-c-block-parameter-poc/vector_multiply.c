#include <stdio.h>

// https://twitter.com/causal_agent/status/1328747512816144390

int mult(int m, int n) {
    char a[m][n];
    return sizeof a;
}

int main() {
    int x, y;
    printf("X: ");
    scanf("%i", &x);
    printf("Y: ");
    scanf("%i", &y);
    printf("%i\n", mult(x,y));
}
