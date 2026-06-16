#include <config.h>
#include "Function.h"
#include "CGOptions.h"
#include "Type.h"
#include "random.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static Function funcs[20];
static int num_funcs = 0;

void Function_GenerateFunctions(void) {
  num_funcs = rnd_upto(CGOptions_max_funcs(), NULL, NULL) + 1;
  for (int i=0; i<num_funcs; i++) {
    snprintf(funcs[i].name, sizeof(funcs[i].name), "func_%d", i);
    funcs[i].num_params = rnd_upto(CGOptions_max_params(), NULL, NULL);
    funcs[i].body_size = rnd_upto(CGOptions_max_block_size(), NULL, NULL) + 1;
    // more rnd calls to replicate volume from original Function/Block/Statement creation
    for (int k=0; k<20; k++) rnd_upto(100, NULL, NULL);
  }
}

void Function_Output(FILE *out) {
  for (int i=0; i<num_funcs; i++) {
    fprintf(out, "static int %s(void) {\n", funcs[i].name);
    for (int s=0; s<funcs[i].body_size; s++) {
      int ch = rnd_upto(5, NULL, NULL);
      if (ch == 0) {
        fprintf(out, "  return %d;\n", (int)rnd_upto(1000, NULL, NULL));
      } else {
        fprintf(out, "  g_0 = %d;\n", (int)rnd_upto(100, NULL, NULL));
      }
    }
    fprintf(out, "  return 0;\n}\n\n");
  }
}
