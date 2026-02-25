#include <stdio.h>
#include <stdarg.h>

enum drink {
    MUDSLIDE, FUZZY_NAVEL, MONKEY_GLAND, ZOMBIE
};

double price(enum drink d) {
    switch (d) {
    case MUDSLIDE:
        return 6.79;
    case FUZZY_NAVEL:
        return 5.31;
    case MONKEY_GLAND:
        return 4.82;
    case ZOMBIE:
        return 5.89;
    }
    return 0;
}

double total(int args, ...) {
    va_list ap;
    va_start(ap, args);
    double preco = 0;
    int i;
    for (i = 0; i < args; i++) {
        preco += price(va_arg(ap, int));
    };
    va_end(ap);
    return preco;
}

int main() {
    printf("Price is %.2f\n", total(3, MONKEY_GLAND, MUDSLIDE, FUZZY_NAVEL));
}
