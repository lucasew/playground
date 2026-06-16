#ifndef FACTMGR_H
#define FACTMGR_H

#include <stddef.h>

struct CGContext;
struct Fact;

typedef struct FactMgr {
  struct Fact *global_facts[100];
  int num_global_facts;
  void *map_facts_in;
  void *map_facts_out;
} FactMgr;

struct FactMgr *get_fact_mgr(struct CGContext *cg);

#endif
