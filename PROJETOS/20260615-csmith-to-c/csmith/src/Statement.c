#include <config.h>
#include "Statement.h"
#include "CGContext.h"
#include "Expression.h"
#include "Block.h"
#include "FactMgr.h"
#include "random.h"
#include "CGOptions.h"
#include "Error.h"
#include "Function.h"
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <string.h>

static int sid = 0;

static void InitProbabilityTable(void) {
  /* probability init; probs via rnd in make for port */
}

static eStatementType StatementProbability(struct CGContext *cg) {
  (void)cg;
  int value = rnd_upto(100, NULL, NULL);
  if (value < 30) return eAssign;
  if (value < 50) return eIfElse;
  if (value < 65) return eFor;
  if (value < 80) return eReturn;
  if (value < 90) return eInvoke;
  return eBlock;
}

Statement *Statement_make_random(struct CGContext *cg_context, eStatementType t) {
  if (!cg_context) {
    Statement *s = (Statement *)malloc(sizeof(Statement));
    if (!s) { fprintf(stderr, "malloc fail\n"); exit(1); }
    memset(s, 0, sizeof(*s));
    s->kind = eAssign;
    return s;
  }
  if (cg_context->expr_depth > CGOptions_max_expr_depth()) {
    t = eReturn;
  }
  if (t == MAX_STATEMENT_TYPE) {
    InitProbabilityTable();
    t = StatementProbability(cg_context);
  }
  Statement *s = (Statement *)malloc(sizeof(Statement));
  if (!s) {
    fprintf(stderr, "malloc fail in Statement_make_random\n");
    exit(1);
  }
  memset(s, 0, sizeof(*s));
  s->kind = t;
  /* pre facts/effect stub */
  cg_context->expr_depth = 0;
  if (t == eBlock || t == eIfElse || t == eFor) {
    cg_context->blk_depth++;
  }
  switch (t) {
  case eAssign:
    s->lhs = Expression_make_random(cg_context, NULL, NULL, false, false, 0);
    s->rhs = Expression_make_random(cg_context, NULL, NULL, false, false, 0);
    break;
  case eIfElse:
    s->cond = Expression_make_random(cg_context, NULL, NULL, false, false, 0);
    s->thenb = Block_make_random(cg_context);
    if (rnd_upto(2, NULL, NULL)) s->elseb = Block_make_random(cg_context);
    break;
  case eFor:
    s->cond = Expression_make_random(cg_context, NULL, NULL, false, false, 0);
    s->thenb = Block_make_random(cg_context);
    break;
  case eReturn:
    s->ret = Expression_make_random(cg_context, NULL, NULL, false, false, 0);
    break;
  case eBlock:
    s->thenb = Block_make_random(cg_context);
    break;
  case eInvoke:
  default:
    s->rhs = Expression_make_random(cg_context, NULL, NULL, false, false, 0);
    s->kind = eInvoke;
    break;
  }
  /* volume and post */
  for (int i = 0; i < 4; i++) rnd_upto(100, NULL, NULL);
  if (t == eBlock || t == eIfElse || t == eFor) {
    cg_context->blk_depth--;
  }
  sid++;
  return s;
}

void Statement_output(Statement *s, FILE *out) {
  if (!s || !out) return;
  switch (s->kind) {
  case eAssign:
    fprintf(out, "  ");
    Expression_output(s->lhs, out);
    fprintf(out, " = ");
    Expression_output(s->rhs, out);
    fprintf(out, ";\n");
    break;
  case eIfElse:
    fprintf(out, "  if (");
    Expression_output(s->cond, out);
    fprintf(out, ") {\n");
    if (s->thenb) Block_output(s->thenb, out);
    fprintf(out, "  }");
    if (s->elseb) {
      fprintf(out, " else {\n");
      Block_output(s->elseb, out);
      fprintf(out, "  }");
    }
    fprintf(out, "\n");
    break;
  case eFor:
    fprintf(out, "  for ( ; ");
    Expression_output(s->cond, out);
    fprintf(out, " ; ) {\n");
    if (s->thenb) Block_output(s->thenb, out);
    fprintf(out, "  }\n");
    break;
  case eReturn:
    fprintf(out, "  return ");
    if (s->ret) Expression_output(s->ret, out);
    fprintf(out, ";\n");
    break;
  case eInvoke:
    fprintf(out, "  ");
    Expression_output(s->rhs, out);
    fprintf(out, ";\n");
    break;
  case eBlock:
    if (s->thenb) Block_output(s->thenb, out);
    break;
  default:
    fprintf(out, "  g_0 = 0;\n");
    break;
  }
}

int Statement_get_blk_depth(Statement *s) { (void)s; return 0; }
