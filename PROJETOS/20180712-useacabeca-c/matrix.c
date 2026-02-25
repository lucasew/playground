#include<stdio.h>
#include<unistd.h>

int main() {
    char c[2];
    while (fgets(c, sizeof(c), stdin)) {
        printf("%s", c);
        usleep(50000);
        fflush(stdout);
    }
    return 0;
    printf("\n");
}
