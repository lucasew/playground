#ifndef EXPRESSION_H
#define EXPRESSION_H

#include <stdio.h>
#include <stdbool.h>

typedef enum {
  eConstant = 0,
  eVariable,
  eFunction,
  eAssignment,
  eCommaExpr,
  MAX_TERM_TYPES
} eTermType;

struct CGContext;
struct Type;
struct CVQualifiers;

typedef struct Expression {
  eTermType kind;
  int value; /* for constant */
  char var_name[32]; /* for var */
  struct Expression *lhs; /* for assign/comma */
  struct Expression *rhs;
  /* more for funcall etc */
} Expression;

Expression *Expression_make_random(struct CGContext *cg_context, const struct Type *type, const struct CVQualifiers *qfer, bool no_func, bool no_const, eTermType tt);
void Expression_output(Expression *e, FILE *out);

#endif
