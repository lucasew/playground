#include <errno.h>
#include <string.h>
#include <stdio.h>

void error(char *err) {
    fprintf(stderr, "%s", err);
    fprintf(stderr, "%s", strerror(errno));
}
