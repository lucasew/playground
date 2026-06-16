#include <string.h>
#include <stdlib.h>
static char *my_strdup(const char *s){if(!s)s="";size_t l=strlen(s)+1;char *p=malloc(l);if(p)memcpy(p,s,l);return p;}
#include "RandomNumber.h"
#include "AbsRndNumGenerator.h"
#include <assert.h>
#include <stdlib.h>
#include <string.h>
static RandomNumber *rnd_instance = NULL;
void RandomNumber_CreateInstance(RNDNUM_GENERATOR rImpl, const unsigned long seed) {
  if (!rnd_instance) {
    rnd_instance = (RandomNumber*)malloc(sizeof(RandomNumber));
    rnd_instance->curr_generator_ = AbsRndNumGenerator_make_rndnum_generator(rImpl, seed);
  }
}
RandomNumber *RandomNumber_GetInstance(void) { return rnd_instance; }
AbsRndNumGenerator *RandomNumber_GetRndNumGenerator(void) { return rnd_instance ? rnd_instance->curr_generator_ : NULL; }
RNDNUM_GENERATOR RandomNumber_SwitchRndNumGenerator(RNDNUM_GENERATOR rImpl) { (void)rImpl; return rDefaultRndNumGenerator; }
void RandomNumber_doFinalization(void) { if (rnd_instance) { free(rnd_instance); rnd_instance = NULL; } }
char *RandomNumber_get_prefixed_name(RandomNumber *r, const char *name) { (void)r; return my_strdup(name?name:""); }
char *RandomNumber_trace_depth(RandomNumber *r) { (void)r; return my_strdup(""); }
void RandomNumber_get_sequence(RandomNumber *r, char **seq) { (void)r; if(seq) *seq=my_strdup(""); }
unsigned int RandomNumber_rnd_upto(RandomNumber *r, unsigned int n, const struct Filter *f, const char *w) {
  (void)r; return AbsRndNumGenerator_rnd_upto(r?r->curr_generator_:NULL, n, f, w);
}
bool RandomNumber_rnd_flipcoin(RandomNumber *r, unsigned int p, const struct Filter *f, const char *w) {
  (void)r; return AbsRndNumGenerator_rnd_flipcoin(r?r->curr_generator_:NULL, p, f, w);
}
char *RandomNumber_RandomHexDigits(RandomNumber *r, int num) {
  (void)r; return AbsRndNumGenerator_RandomHexDigits(r?r->curr_generator_:NULL, num);
}
char *RandomNumber_RandomDigits(RandomNumber *r, int num) {
  (void)r; return AbsRndNumGenerator_RandomDigits(r?r->curr_generator_:NULL, num);
}
