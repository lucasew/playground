#ifndef UTIL_H
#define UTIL_H
/* stub */
typedef int intvec;
char *gensym(const char *basename);
int *permute(int *in, int len);
int expand_within_ranges(const unsigned int *ranges, int nr, int *out);
#endif
