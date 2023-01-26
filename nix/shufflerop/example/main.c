#include<stdio.h>

extern const int a;
extern const int b;

int main() {
    printf("Sum: %i + %i = %i\n", a, b, a + b);
    printf("Ptr: a=%p b=%p\n", &a, &b);
};
