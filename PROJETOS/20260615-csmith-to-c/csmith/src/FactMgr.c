#include <config.h>
#include "FactMgr.h"
#include "CGContext.h"
#include <stdlib.h>

static struct FactMgr fm_instance;

struct FactMgr *get_fact_mgr(struct CGContext *cg) {
  (void)cg;
  if (fm_instance.num_global_facts == 0) {
    fm_instance.num_global_facts = 0;
    /* init more if needed */
  }
  return &fm_instance;
}
