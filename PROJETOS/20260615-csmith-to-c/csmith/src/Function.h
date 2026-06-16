#ifndef Function_H
#define Function_H
#include <stdio.h>

typedef struct Function {
  char name[32];
  int num_params;
  int body_size;
} Function;

void Function_GenerateFunctions(void);
void Function_Output(FILE *out);
#endif
