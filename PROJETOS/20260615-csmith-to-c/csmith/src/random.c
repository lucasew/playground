#include <config.h>
#define _GNU_SOURCE
#include "random.h"
#include "RandomNumber.h"
#include "AbsRndNumGenerator.h"
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
static char *my_strdup_local(const char *s){ if(!s)s=""; size_t l=strlen(s)+1; char *p=malloc(l); if(p) memcpy(p,s,l); return p; }

unsigned int rnd_upto(const unsigned int n, const struct Filter *f, const char *where) {
  RandomNumber *rn = RandomNumber_GetInstance();
  if (rn) return RandomNumber_rnd_upto(rn, n, f, where);
  if (n == 0) return 0;
  return (unsigned int)(rand() % n);
}

int rnd_flipcoin(const unsigned int p, const struct Filter *f, const char *where) {
  RandomNumber *rn = RandomNumber_GetInstance();
  if (rn) return RandomNumber_rnd_flipcoin(rn, p, f, where);
  if (p >= 100) return 1;
  return (rand() % 100) < (int)p;
}

char *PureRandomHexDigits(int num) { return RandomHexDigits(num); }
char *PureRandomDigits(int num) { return RandomDigits(num); }

unsigned int pure_rnd_upto(const unsigned int n, const struct Filter *f, const char *where) {
  return rnd_upto(n, f, where);
}

int pure_rnd_flipcoin(const unsigned int p, const struct Filter *f, const char *where) {
  return rnd_flipcoin(p, f, where);
}

char *RandomHexDigits(int num) {
  RandomNumber *rn = RandomNumber_GetInstance();
  if (rn) return RandomNumber_RandomHexDigits(rn, num);
  char *s = malloc(num+1); if(!s) return NULL;
  for(int i=0; i<num; i++) s[i] = "0123456789ABCDEF"[rand()%16];
  s[num]=0; return s;
}

char *RandomDigits(int num) {
  RandomNumber *rn = RandomNumber_GetInstance();
  if (rn) return RandomNumber_RandomDigits(rn, num);
  char *s = malloc(num+1); if(!s) return NULL;
  for(int i=0; i<num; i++) s[i] = "0123456789"[rand()%10];
  s[num]=0; return s;
}

char *get_prefixed_name(const char *name) {
  (void)name;
  /* simple for now; real would use RandomNumber */
  return my_strdup_local(name ? name : "");
}

char *trace_depth(void) {
  return my_strdup_local("");
}

void get_sequence(char **sequence) {
  if (sequence) *sequence = my_strdup_local("");
}
