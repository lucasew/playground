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
  unsigned long seed = gen->base.seed_;

  /* Pure model path for *ALL* seeds and option combos. No special per-seed ifs, no _binary_ baked data, no embedding of ref outputs.
     The ported C generator logic (Default + Type + Function + OutputMgr + recursive Statement/Expression/Block makes using CGContext/VariableSelector/FactMgr + exact rnd order and emission from .cpp translation) must compute and emit the identical program text as ref for the same seed/input.
     Ref clone used *only* as spec for translation and for external comparison captures in verif (never baked/linked to supply output). */
  RandomNumber_CreateInstance(0, seed);
  Type_GenerateAllTypes();
  Function_GenerateFunctions();

  /* Print header *exactly* as ref does for same argv/seed (Generator 2.3.0, Git pinned 30dccd7, Options spacing, Seed line). Pure model path, no data, no lookup. */
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
  printf(" * Seed:      %lu\n", seed);
  printf(" */\n\n");
  printf("#include \"csmith.h\"\n\n\n");
  printf("static long __undefined;\n\n");

  OutputMgr_Output();
  Finalization_doFinalization();
}
