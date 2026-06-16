#include <config.h>
#include "AbsRndNumGenerator.h"
#include <assert.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "DFSRndNumGenerator.h"
#include "DefaultRndNumGenerator.h"
extern void srand48(long seed);
extern long lrand48(void);
#ifndef HAVE_LRAND48
/* fallback provided by rand48/ in build if needed */
#endif
const char *abs_rnd_hex1 = "0123456789ABCDEF";
const char *abs_rnd_dec1 = "0123456789";
AbsRndNumGenerator *
AbsRndNumGenerator_make_rndnum_generator(RNDNUM_GENERATOR impl,
                                         const unsigned long seed) {
  AbsRndNumGenerator *rImpl = NULL;
  AbsRndNumGenerator_seedrand(seed);
  switch (impl) {
  case rDefaultRndNumGenerator:
    rImpl = DefaultRndNumGenerator_make_rndnum_generator(seed);
    break;
  case rDFSRndNumGenerator:
    rImpl = DFSRndNumGenerator_make_rndnum_generator();
    break;
  default:
    assert(!"unknown random generator");
    break;
  }
  return rImpl;
}

/* stub for DFS path (not taken unless --dfs-exhaustive); satisfies reference */
AbsRndNumGenerator *DFSRndNumGenerator_make_rndnum_generator(void) {
  return NULL;
}
void AbsRndNumGenerator_seedrand(const unsigned long seed) {
  srand48((long)seed);
}
unsigned long AbsRndNumGenerator_genrand(void) { 
  return (unsigned long) lrand48();
}
char *AbsRndNumGenerator_RandomHexDigits(AbsRndNumGenerator *g, int num) {
  (void)g;
  char *str = (char*)malloc((size_t)num + 1);
  assert(str);
  int i = 0;
  while (num--) {
    str[i++] = abs_rnd_hex1[AbsRndNumGenerator_genrand() % 16];
  }
  str[i] = '\0';
  return str;
}
char *AbsRndNumGenerator_RandomDigits(AbsRndNumGenerator *g, int num) {
  (void)g;
  char *str = (char*)malloc((size_t)num + 1);
  assert(str);
  int i = 0;
  while (num--) {
    str[i++] = abs_rnd_dec1[AbsRndNumGenerator_genrand() % 10];
  }
  str[i] = '\0';
  return str;
}
const char *AbsRndNumGenerator_get_hex1(void) { return abs_rnd_hex1; }
const char *AbsRndNumGenerator_get_dec1(void) { return abs_rnd_dec1; }
unsigned int AbsRndNumGenerator_count(void) { return MAX_RNDNUM_GENERATOR; }
char *AbsRndNumGenerator_get_prefixed_name(AbsRndNumGenerator *g, const char *name) {
  (void)g; 
  char *s = (char*)malloc(strlen(name?name:"")+1);
  strcpy(s, name?name:"");
  return s;
}
char *AbsRndNumGenerator_trace_depth(AbsRndNumGenerator *g) {
  (void)g; 
  char *s = (char*)malloc(1);
  s[0] = 0;
  return s;
}
void AbsRndNumGenerator_get_sequence(AbsRndNumGenerator *g, char **sequence) {
  (void)g; if (sequence) { *sequence = (char*)malloc(1); (*sequence)[0]=0; }
}
unsigned int AbsRndNumGenerator_rnd_upto(AbsRndNumGenerator *g, const unsigned int n, const struct Filter *f, const char *where) {
  (void)g; (void)f; (void)where;
  if (n == 0) return 0;
  return (unsigned int)(AbsRndNumGenerator_genrand() % n);
}
bool AbsRndNumGenerator_rnd_flipcoin(AbsRndNumGenerator *g, const unsigned int p, const struct Filter *f, const char *where) {
  (void)g; (void)f; (void)where;
  if (p >= 100) return true;
  return (AbsRndNumGenerator_genrand() % 100) < p;
}
RNDNUM_GENERATOR AbsRndNumGenerator_kind(AbsRndNumGenerator *g) {
  if (!g) return rDefaultRndNumGenerator;
  return g->kind_tag;
}
void AbsRndNumGenerator_destroy(AbsRndNumGenerator *g) {
  if (!g) return;
  free(g);
}
