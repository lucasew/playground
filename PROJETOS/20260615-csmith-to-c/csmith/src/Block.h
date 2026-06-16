#ifndef Block_H
#define Block_H

#include <stdio.h>

struct Statement;
struct CGContext;

typedef struct Block {
  struct Statement **stms;
  int num;
  struct Block *parent;
} Block;

Block *Block_make_random(struct CGContext *cg_context);
void Block_output(Block *b, FILE *out);
Block *Block_make_dummy_block(struct CGContext *cg_context);

#endif
