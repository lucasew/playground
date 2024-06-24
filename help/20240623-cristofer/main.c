#include<stdio.h>
#include<stdlib.h>

int main() {
  printf("hello, world\n");

  int v = 2;
  char* teste = NULL;
  char* texto = "bem loco";

  printf("valor: %p\n", &v);
  printf("texto: %s\n", texto);
  printf("texto: %s\n", teste);
  printf("int: %i\n", *teste);
  return 0;
}
