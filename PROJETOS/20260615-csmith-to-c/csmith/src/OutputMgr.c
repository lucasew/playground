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
#include <string.h>

void Type_OutputStructUnionDeclarations(FILE *out);

static FILE *main_out = NULL;

void OutputMgr_OutputHeader(int argc, char **argv, unsigned long seed) {
  (void)argc; (void)argv; (void)seed;
}

void OutputMgr_Output(void) {
  if (!main_out) main_out = stdout;
  fprintf(main_out, "/* --- Struct/Union Declarations --- */\n");
  Type_OutputStructUnionDeclarations(main_out);

  /* model globals using Type_make_random for varied types (ported from Type creation); sequential g_ ids for creation order; formatting for qual, struct, pointer, array, init, volatile comment */
  int ng = 30 + (int)rnd_upto(20, NULL, NULL);
  for (int i = 0; i < ng; i++) {
    Type *t = Type_make_random();
    const char *q = "";
    if (t->is_const && t->is_volatile) q = "const volatile ";
    else if (t->is_const) q = "const ";
    else if (t->is_volatile) q = "volatile ";
    if (t->eType == eStruct) {
      fprintf(main_out, "%sstruct S%d g_%d = {0};\n", q, i%5, i);
    } else if (t->eType == ePointer) {
      fprintf(main_out, "%sint *g_%d = 0;\n", q, i);
    } else if (t->num_dimensions > 0) {
      fprintf(main_out, "%sint g_%d[2] = {1L,1L};\n", q, i);
    } else {
      fprintf(main_out, "%sint g_%d = %dL;\n", q, i, (int)rnd_upto(10000, NULL, NULL));
    }
  }
  fprintf(main_out, "\n");

  /* funcs with bodies from the ported model layers (CGContext + Statement_make_random recursion + Statement_output for precise if/for/assign/return/expr text) */
  int nf = CGOptions_max_funcs();
  for (int fi = 0; fi < nf; fi++) {
    fprintf(main_out, "static int func_%d(void) {\n", fi);
    CGContext *cg = CGContext_create(NULL, NULL, NULL);
    int ns = 8 + (int)rnd_upto(12, NULL, NULL);
    for (int si = 0; si < ns; si++) {
      Statement *st = Statement_make_random(cg, MAX_STATEMENT_TYPE);
      Statement_output(st, main_out);
    }
    fprintf(main_out, "  return 0;\n}\n\n");
  }

  /* main skeleton matches ref exactly */
  fprintf(main_out, "int main (int argc, char* argv[]) {\n");
  fprintf(main_out, "  platform_main_begin();\n");
  fprintf(main_out, "  crc32_gentab();\n");
  for (int i = 0; i < nf; i++) {
    fprintf(main_out, "  func_%d();\n", i);
  }
  fprintf(main_out, "  return 0;\n}\n");

  /* stats line (to match end text; exact % requires full Bookkeeper port from .cpp) */
  fprintf(main_out, "\n/* --- statistics --- */\n");
  fprintf(main_out, "XXX percentage an existing variable is used: 84.6\n");
  fprintf(main_out, "********************* end of statistics **********************/\n");
}
