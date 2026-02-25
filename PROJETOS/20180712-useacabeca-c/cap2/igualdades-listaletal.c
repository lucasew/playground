#include <stdio.h>

int main() {
    int doses[] = {1, 3, 2, 1000};
    printf("%i\n", doses[3]);
    printf("%i\n", *(doses + 3));
    printf("%i\n", *(3 + doses));
    printf("%i\n", 3[doses]);
    // todos retornam 1000
}
