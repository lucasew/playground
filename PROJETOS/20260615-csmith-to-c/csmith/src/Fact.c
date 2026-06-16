#include <config.h>
#include "Fact.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

Fact *Fact_make(void) {
  Fact *f = (Fact *)malloc(sizeof(Fact));
  if (!f) { fprintf(stderr, "malloc fail Fact\n"); exit(1); }
  memset(f, 0, sizeof(*f));
  f->d = 0; /* match the port's minimal Fact {int d;} in Fact.h */
  return f;
}

void Fact_destroy(Fact *f) { if (f) free(f); }

