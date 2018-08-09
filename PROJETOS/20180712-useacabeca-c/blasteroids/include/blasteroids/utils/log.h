#ifndef _BLASTEROIDS_UTILS_LOG
#define _BASTEROIDS_UTILS_LOG
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <errno.h>
/*
   void debug(char *msg, ...);

   void info(char *msg, ...);

   void warning(char *msg, ...);

   void error(char *msg);
   */
#define _print(prefix, ...) do { \
    fprintf(stderr, "%s: ", prefix); \
    fprintf(stderr, __VA_ARGS__); \
    fprintf(stderr, "\n"); \
} while(0)

#define debug(...) \
    _print("DEBG", __VA_ARGS__)

#define info(...) \
    _print("INFO", __VA_ARGS__)

#define warning(...)  \
    _print("WARN", __VA_ARGS__)

#define error(msg) do {\
    _print("ERR", "%s: %s", msg, strerror(errno)); \
    raise(SIGINT); \
    } while(0)

#endif
