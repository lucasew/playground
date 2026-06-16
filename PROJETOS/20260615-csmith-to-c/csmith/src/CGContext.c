#include <config.h>
#include "CGContext.h"
#include "Function.h"
#include "Effect.h"
#include "Block.h"
#include <stdlib.h>
#include <string.h>

CGContext *CGContext_create(struct Function *current_func, struct Effect *eff_context, struct Effect *eff_accum) {
  CGContext *c = (CGContext *)malloc(sizeof(CGContext));
  if (!c) {
    fprintf(stderr, "malloc fail in CGContext_create\n");
    exit(1);
  }
  memset(c, 0, sizeof(*c));
  c->current_func = current_func;
  c->blk_depth = 0;
  c->expr_depth = 0;
  c->flags = 0;
  c->curr_blk = NULL;
  c->effect_context = eff_context;
  c->effect_accum = eff_accum;
  c->rw_directive = 0;
  c->iv_bounds = 0;
  c->curr_rhs = 0;
  c->call_chain = 0;
  c->effect_stm = 0;
  return c;
}

struct Function *CGContext_get_current_func(CGContext *c) {
  return c ? c->current_func : NULL;
}

struct Block *CGContext_get_current_block(CGContext *c) {
  return c ? c->curr_blk : NULL;
}

void CGContext_set_current_block(CGContext *c, struct Block *b) {
  if (c) c->curr_blk = b;
}

int CGContext_get_blk_depth(CGContext *c) {
  return c ? c->blk_depth : 0;
}

int CGContext_get_expr_depth(CGContext *c) {
  return c ? c->expr_depth : 0;
}
