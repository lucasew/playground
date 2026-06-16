#include <config.h>
#include "VariableSelector.h"
#include "CGContext.h"
#include "Type.h"
#include "random.h"
#include <stdlib.h>

struct Variable *VariableSelector_select(struct CGContext *cg, const struct Type *type, bool is_volatile, bool is_const, bool is_addr_taken, bool is_dereference, int indirect_level) {
  (void)cg; (void)type; (void)is_volatile; (void)is_const; (void)is_addr_taken; (void)is_dereference; (void)indirect_level;
  /* simple select using rnd; real would use visible vars list + facts */
  /* for volume, just return stub or use global g_0 style */
  return NULL; /* caller falls back to g_ or int */
}
