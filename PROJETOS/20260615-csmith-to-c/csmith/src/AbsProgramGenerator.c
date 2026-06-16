#include <config.h>
#include "AbsProgramGenerator.h"
#include "CGOptions.h"
#include "DefaultProgramGenerator.h"
#include "RandomNumber.h"
#include "platform.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void CGOptions_set_default_settings(void);
void platform_init(void);

static AbsProgramGenerator *current_generator_ = NULL;

AbsProgramGenerator *AbsProgramGenerator_CreateInstance(int argc, char *argv[], unsigned long seed) {
  // ported from original: parse options into CGOptions (simplified for required + key)
  for (int i=1; i<argc; i++) {
    if (strcmp(argv[i], "--no-arrays") == 0) CGOptions_arrays_set(false);
    else if (strcmp(argv[i], "--arrays") == 0) CGOptions_arrays_set(true);
    else if (strcmp(argv[i], "--no-structs") == 0) CGOptions_use_struct_set(false);
    else if (strcmp(argv[i], "--structs") == 0) CGOptions_use_struct_set(true);
    else if (strcmp(argv[i], "--no-unions") == 0) CGOptions_use_union_set(false);
    else if (strcmp(argv[i], "--unions") == 0) CGOptions_use_union_set(true);
    else if (strcmp(argv[i], "--no-bitfields") == 0) CGOptions_bitfields_set(false);
    else if (strcmp(argv[i], "--bitfields") == 0) CGOptions_bitfields_set(true);
    else if (strcmp(argv[i], "--max-funcs") == 0 && i+1<argc) {
      int v; if (sscanf(argv[++i], "%d", &v)==1) CGOptions_max_funcs_set(v);
    }
    // port more from original RandomProgramGenerator.cpp parse for full
  }
  CGOptions_set_default_settings();
  // platform etc.
  platform_init();
  RandomNumber_CreateInstance(0 /* rDefault */, seed);
  DefaultProgramGenerator *dgen = DefaultProgramGenerator_Create(argc, argv, seed);
  current_generator_ = (AbsProgramGenerator *)dgen;
  return current_generator_;
}

void AbsProgramGenerator_goGenerator(AbsProgramGenerator *gen) {
  // delegate to Default (layout: Default* cast from Abs* which points to embedded base)
  DefaultProgramGenerator *d = (DefaultProgramGenerator *)gen;
  DefaultProgramGenerator_goGenerator(d);
}

void AbsProgramGenerator_doFinalization(void) {
  if (current_generator_) free(current_generator_);
  current_generator_ = NULL;
}
