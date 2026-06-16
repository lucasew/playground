#ifndef STATEMENT_H
#define STATEMENT_H

#include <stdio.h>
#include <stdbool.h>

typedef enum {
  eAssign = 0,
  eBlock,
  eFor,
  eIfElse,
  eInvoke,
  eReturn,
  eBreak,
  eContinue,
  eGoto,
  eArrayOp,
  MAX_STATEMENT_TYPE
} eStatementType;

struct Function;
struct Block;
struct Expression;
struct CGContext;

typedef struct Statement {
  eStatementType kind;
  struct Function *func;
  struct Block *parent;
  struct Expression *lhs;
  struct Expression *rhs;
  struct Expression *cond;
  struct Block *thenb;
  struct Block *elseb;
  struct Expression *ret;
} Statement;

Statement *Statement_make_random(struct CGContext *cg_context, eStatementType t);
void Statement_output(Statement *s, FILE *out);
int Statement_get_blk_depth(Statement *s);

#endif
