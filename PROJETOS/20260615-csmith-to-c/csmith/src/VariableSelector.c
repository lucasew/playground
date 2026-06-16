#include <config.h>
#include "VariableSelector.h"
#include "CGContext.h"
#include "Type.h"
#include "random.h"
#include <stdlib.h>
#include <string.h>

struct Variable *VariableSelector_select(struct CGContext *cg, const struct Type *type, bool is_volatile, bool is_const, bool is_addr_taken, bool is_dereference, int indirect_level) {
  (void)cg; (void)type; (void)is_volatile; (void)is_const; (void)is_addr_taken; (void)is_dereference; (void)indirect_level;
  /* guard + stub: return non-null sentinel (no sizeof/incomplete struct Variable issues) to avoid null deref/crash in CGContext/Fact/VarSel/Statement/Expression/Block make paths for normal runs.
     (full Variable is C++ defined; port uses name strings directly in expr/stmt emission) */
  static char sentinel[64];
  memset(sentinel, 0, sizeof(sentinel));
  snprintf(sentinel, sizeof(sentinel), "g_%d", (int)rnd_upto(50, NULL, NULL));
  return (struct Variable *)sentinel;
}
