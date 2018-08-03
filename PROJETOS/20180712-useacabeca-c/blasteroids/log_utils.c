#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void _print(char *prefix, char *msg) {
    fprintf(stderr, "%s: %s\n", prefix, msg);
}

void debug(char *msg) {
    _print("DEBG", msg);
}

void info(char *msg) {
    _print("INFO", msg);
}

void warning(char *msg) {
    _print("WARN", msg);
}

void error(char *msg) {
    char *text;
    sprintf(text, "%s: %s", msg, strerror(errno));
    _print("ERR", text);
    raise(SIGINT);
}
