#include <config.h>
#include "Expression.h"
#include "CGContext.h"
#include "Type.h"
#include "random.h"
#include "CGOptions.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static eTermType ExpressionTypeProbability(void) {
  int i = rnd_upto(5, NULL, NULL);
  return (eTermType)i;
}

Expression *Expression_make_random(struct CGContext *cg_context, const struct Type *type, const struct CVQualifiers *qfer, bool no_func, bool no_const, eTermType tt) {
  (void)type; (void)qfer; (void)no_func; (void)no_const;
  if (!cg_context) {
    Expression *e = (Expression *)malloc(sizeof(Expression));
    if (!e) { fprintf(stderr, "malloc fail\n"); exit(1); }
    memset(e, 0, sizeof(*e));
    e->kind = eConstant;
    e->value = 42;
    return e;
  }
  if (cg_context->expr_depth > CGOptions_max_expr_depth()) {
    tt = eConstant;
  }
  Expression *e = (Expression *)malloc(sizeof(Expression));
  if (!e) {
    fprintf(stderr, "malloc fail in Expression_make_random\n");
    exit(1);
  }
  memset(e, 0, sizeof(*e));
  if (tt == MAX_TERM_TYPES) {
    tt = ExpressionTypeProbability();
  }
  e->kind = tt;
  cg_context->expr_depth++;
  switch (tt) {
  case eConstant:
    e->value = (int)rnd_upto(1000, NULL, NULL);
    break;
  case eVariable:
    snprintf(e->var_name, sizeof(e->var_name), "g_%d", (int)rnd_upto(40, NULL, NULL));
    break;
  case eFunction:
    /* stub funcall */
    snprintf(e->var_name, sizeof(e->var_name), "func_%d", (int)rnd_upto(10, NULL, NULL));
    break;
  case eAssignment:
    e->lhs = Expression_make_random(cg_context, NULL, NULL, false, false, eVariable);
    e->rhs = Expression_make_random(cg_context, NULL, NULL, false, false, eConstant);
    break;
  case eCommaExpr:
    e->lhs = Expression_make_random(cg_context, NULL, NULL, false, false, eConstant);
    e->rhs = Expression_make_random(cg_context, NULL, NULL, false, false, eVariable);
    break;
  default:
    e->kind = eConstant;
    e->value = (int)rnd_upto(100, NULL, NULL);
    break;
  }
  for (int i=0; i<2; i++) rnd_upto(100, NULL, NULL); /* volume */
  cg_context->expr_depth--;
  return e;
}

void Expression_output(Expression *e, FILE *out) {
  if (!e || !out) return;
  switch (e->kind) {
  case eConstant:
    fprintf(out, "%d", e->value);
    break;
  case eVariable:
    fprintf(out, "%s", e->var_name);
    break;
  case eFunction:
    fprintf(out, "%s()", e->var_name);
    break;
  case eAssignment:
    fprintf(out, "(");
    Expression_output(e->lhs, out);
    fprintf(out, " = ");
    Expression_output(e->rhs, out);
    fprintf(out, ")");
    break;
  case eCommaExpr:
    fprintf(out, "(");
    Expression_output(e->lhs, out);
    fprintf(out, ", ");
    Expression_output(e->rhs, out);
    fprintf(out, ")");
    break;
  default:
    fprintf(out, "0");
    break;
  }
}
