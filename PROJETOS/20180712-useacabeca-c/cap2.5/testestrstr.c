#include <stdio.h>
#include <string.h>

int main () {
    char *procura = strstr("fjfijefiojeiofjweofijweoqoqejfiweofej", "eoq");
    if (procura) {
        printf("Eoq foi encontradoi em %p!\n", procura);
        puts(procura);
    }
}
