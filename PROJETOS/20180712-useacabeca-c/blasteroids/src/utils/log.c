#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
/*
void _print(char *prefix, char *msg, ...) {
    va_list args;
    va_start(args, msg);
    fprintf(stderr, "%s: ", prefix);
    vfprintf(stderr, msg, args);
    fprintf(stderr, "\n");
    va_end(args);
}

void debug(char *msg, ...) {
    _print("DEBG", msg, _VA_LIST_);
}

void info(char *msg, ...) {
    va_list args;
    va_start(args, msg);
    _print("INFO", msg, args);
    va_end(args);
}

void warning(char *msg, ...) {
    va_list args;
    va_start(args, msg);
    _print("WARN", msg, args);
    va_end(args);
}

void error(char *msg, ...) {
    va_list args;
    char *text;
    sprintf(text, "%s: %s", msg, strerror(errno));
    _print("ERR", text);
    raise(SIGINT);
}
*/

