#ifndef _BLASTEROIDS_UTILS_SIGNAL
#define _BLASTEROIDS_UTILS_SIGNAL

#include <signal.h>

int catch_signal(int sig, void(*handler)(int));

#endif
