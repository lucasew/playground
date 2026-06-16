#include <config.h>
#include "util.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
static char *my_strdup(const char *s){if(!s)s="";size_t l=strlen(s)+1;char *p=malloc(l);if(p)memcpy(p,s,l);return p;} char *gensym(const char *basename) { static int c=0; char buf[128]; snprintf(buf,128,"%s%d",basename,c++); return my_strdup(buf); }
