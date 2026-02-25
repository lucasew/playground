#include <stdio.h>
#include <limits.h>
#include <float.h>

int main () {
    printf("Valor de INT_MAX é %i\n", INT_MAX);
    printf("Valor de INT_MIN é %i\n", INT_MIN);
    printf("Um int usa %zu bytes\n\n", sizeof(int));

    printf("O valor de FLT_MAX é %f\n", FLT_MAX);
    printf("O valor de FLT_MIN é  %.50f\n", FLT_MIN);
    printf("Um float ocupa %zu bytes\n\n", sizeof(float));
    
    printf("O valor de CHAR_MAX é %d\n", CHAR_MAX); // Eu tambêm esperava uma letra :/
    printf("O valor de CHAR_MIN é %d\n", CHAR_MIN);
    printf("Um char ocupa %zu bytes\n\n", sizeof(char));

    printf("O valor de DBL_MAX é %f\n", DBL_MAX);
    printf("Um double ocupa %zu bytes\n\n", sizeof(double));

    printf("O valor de SHRT_MAX é %hi\n", SHRT_MAX);
    printf("O valor de SHRT_MIN é %hi\n", SHRT_MIN);
    printf("Um short ocupa %zu bytes\n\n", sizeof(short));

    printf("O valor de LONG_MAX é %ld\n", LONG_MAX);
    printf("O valor de LONG_MIN é %ld\n", LONG_MIN);
    printf("Um long ocupa %zu bytes\n\n", sizeof(short));
    return 0;
}
