#ifndef FILTER_H
#define FILTER_H
#include <stdbool.h>
typedef struct Filter { int k; } Filter;
Filter *Filter_new(void);
bool Filter_filter(const Filter *f, int v);
void Filter_enable(Filter *f, int kind);
void Filter_disable(Filter *f, int kind);
#endif
