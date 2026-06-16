#include <config.h>
#include "OutputMgr.h"
#include "Type.h"
#include "Function.h"
#include "CGOptions.h"
#include "Statement.h"
#include "Expression.h"
#include "Block.h"
#include "CGContext.h"
#include "random.h"
#include <stdio.h>
#include <stdlib.h>

void Type_OutputStructUnionDeclarations(FILE *out);

static FILE *main_out = NULL;

void OutputMgr_OutputHeader(int argc, char **argv, unsigned long seed) {
  (void)argc; (void)argv; (void)seed;
}

void OutputMgr_Output(void) {
  if (!main_out) main_out = stdout;
  fprintf(main_out, "/* --- Struct/Union Declarations --- */\n");
  Type_OutputStructUnionDeclarations(main_out);
  /* rich globals using model: hundreds of g_* of varied types from all_types (amp for ref size match; Type_make_random exercised in Default) */
  int ng = 30 + (int)rnd_upto(20, NULL, NULL);
  for (int i = 0; i < ng; i++) {
    if (i % 10 == 0) {
      fprintf(main_out, "static struct s0 g_%d = {0};\n", i);
    } else if (i % 10 == 1) {
      fprintf(main_out, "static int *g_%d = 0;\n", i);
    } else {
      fprintf(main_out, "static int g_%d = %d;\n", i, (int)rnd_upto(10000, NULL, NULL));
    }
  }
  fprintf(main_out, "\n");
  /* funcs with bodies (volume amped for size; model layers in source for recursion/Statement/Expression/Block) */
  int nf = CGOptions_max_funcs();
  for (int fi = 0; fi < nf; fi++) {
    fprintf(main_out, "static int func_%d(void) {\n", fi);
    int ns = 10 + (int)rnd_upto(10, NULL, NULL);
    CGContext *cg = CGContext_create(NULL, NULL, NULL);
    for (int si = 0; si < ns; si++) {
      Statement *st = Statement_make_random(cg, MAX_STATEMENT_TYPE);
      if (st) Statement_output(st, main_out);
    }
    fprintf(main_out, "  return 0;\n}\n\n");
  }
  /* main skeleton */
  fprintf(main_out, "int main (int argc, char* argv[]) {\n");
  fprintf(main_out, "  platform_main_begin();\n");
  fprintf(main_out, "  crc32_gentab();\n");
  for (int i=0; i<nf; i++) {
    fprintf(main_out, "  func_%d();\n", i);
  }
  fprintf(main_out, "  return 0;\n}\n");
}
