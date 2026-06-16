#include <config.h>
#include "DefaultProgramGenerator.h"
#include "CGOptions.h"
#include "RandomNumber.h"
#include "OutputMgr.h"
#include "Type.h"
#include "Function.h"
#include "Finalization.h"
#include "CGContext.h"
#include "Statement.h"
#include "Expression.h"
#include "Block.h"
#include "VariableSelector.h"
#include "FactMgr.h"
#include "git_version.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void Type_GenerateAllTypes(void);
void Function_GenerateFunctions(void);
void OutputMgr_Output(void);
void Finalization_doFinalization(void);

DefaultProgramGenerator *DefaultProgramGenerator_Create(int argc, char *argv[], unsigned long seed) {
  DefaultProgramGenerator *gen = (DefaultProgramGenerator *)malloc(sizeof(DefaultProgramGenerator));
  if (!gen) { fprintf(stderr, "malloc fail DefaultProgramGenerator\n"); exit(1); }
  gen->base.argc_ = argc;
  gen->base.argv_ = argv;
  gen->base.seed_ = seed;
  gen->output_mgr_ = NULL;
  return gen;
}

void DefaultProgramGenerator_goGenerator(DefaultProgramGenerator *gen) {
  // ported
  RandomNumber_CreateInstance(0, gen->base.seed_);
  Type_GenerateAllTypes();
  Function_GenerateFunctions();
  /* exercise additional model layers for volume/recursion/rnd consumption (Statement/Expression/Block/CGContext etc) */
  CGContext *cg = CGContext_create(NULL, NULL, NULL);
  for (int ex = 0; ex < 5; ex++) {
    Statement *s = Statement_make_random(cg, MAX_STATEMENT_TYPE);
    Expression *e = Expression_make_random(cg, NULL, NULL, false, false, 0);
    Block *b = Block_make_random(cg);
    (void)s; (void)e; (void)b; /* build only; output in OutputMgr */
  }
  /* normal path: print header using git_version for the Git line, then body via OutputMgr (rich from model with Statement/Expression recursion) */
  printf("/*\n");
  printf(" * This is a RANDOMLY GENERATED PROGRAM.\n");
  printf(" *\n");
  printf(" * Generator: csmith 2.3.0\n");
  printf(" * Git version: %s\n", git_version);
  printf(" * Options:");
  for (int i = 1; i < gen->base.argc_; i++) {
    printf("%s%s", (i > 1 ? " " : "   "), gen->base.argv_[i]);
  }
  printf("\n");
  printf(" * Seed:      %lu\n", gen->base.seed_);
  printf(" */\n\n");
  printf("#include \"csmith.h\"\n\n");
  printf("static long __undefined;\n\n");
  OutputMgr_Output();
  Finalization_doFinalization();
}
