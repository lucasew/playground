#ifndef ERROR_H
#define ERROR_H
#include <stdio.h>
void Error(int e, const char *fmt, ...);
int Error_get_error(void);
#endif
