#include <config.h>
#include "DefaultRndNumGenerator.h"
#include "AbsRndNumGenerator.h"
#include <stdlib.h>
#include <assert.h>
AbsRndNumGenerator *DefaultRndNumGenerator_make_rndnum_generator(const unsigned long seed) {
  (void)seed;
  AbsRndNumGenerator *g = (AbsRndNumGenerator*)malloc(sizeof(AbsRndNumGenerator));
  g->kind_tag = rDefaultRndNumGenerator;
  g->impl_data = NULL;
  return g;
}
