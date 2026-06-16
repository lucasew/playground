#include <config.h>
#include "Block.h"
#include "Statement.h"
#include "CGContext.h"
#include "CGOptions.h"
#include "random.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

Block *Block_make_random(struct CGContext *cg_context) {
  if (!cg_context) {
    Block *b = (Block *)malloc(sizeof(Block));
    if (!b) { fprintf(stderr, "malloc fail\n"); exit(1); }
    b->stms = NULL;
    b->num = 0;
    b->parent = NULL;
    return b;
  }
  Block *b = (Block *)malloc(sizeof(Block));
  if (!b) {
    fprintf(stderr, "malloc fail in Block_make_random\n");
    exit(1);
  }
  int n = rnd_upto(CGOptions_max_block_size(), NULL, NULL) + 1;
  b->stms = (struct Statement **)malloc(n * sizeof(struct Statement *));
  if (!b->stms) {
    fprintf(stderr, "malloc fail stms\n");
    exit(1);
  }
  b->num = n;
  b->parent = NULL;
  for (int i = 0; i < n; i++) {
    b->stms[i] = Statement_make_random(cg_context, MAX_STATEMENT_TYPE);
  }
  for (int i = 0; i < 3; i++) rnd_upto(100, NULL, NULL); /* volume from original */
  return b;
}

Block *Block_make_dummy_block(struct CGContext *cg_context) {
  (void)cg_context;
  return Block_make_random(cg_context);
}

void Block_output(Block *b, FILE *out) {
  if (!b || !out) return;
  for (int i = 0; i < b->num; i++) {
    Statement_output(b->stms[i], out);
  }
}
