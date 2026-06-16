#ifndef CGCONTEXT_H
#define CGCONTEXT_H

#include <stdbool.h>

struct Function;
struct Block;
struct Effect;

typedef struct CGContext {
  struct Function *current_func;
  int blk_depth;
  int expr_depth;
  int flags;
  struct Block *curr_blk;
  struct Effect *effect_context;
  struct Effect *effect_accum;
  /* more from .cpp for full init */
  void *rw_directive;
  int iv_bounds;
  void *curr_rhs;
  void *call_chain;
  void *effect_stm;
} CGContext;

CGContext *CGContext_create(struct Function *current_func, struct Effect *eff_context, struct Effect *eff_accum);
struct Function *CGContext_get_current_func(CGContext *c);
struct Block *CGContext_get_current_block(CGContext *c);
void CGContext_set_current_block(CGContext *c, struct Block *b);
int CGContext_get_blk_depth(CGContext *c);
int CGContext_get_expr_depth(CGContext *c);

#endif
