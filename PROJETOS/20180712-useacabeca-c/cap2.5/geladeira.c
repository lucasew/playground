#include <stdio.h>
#include <string.h>

void print_reverse(char *s) {
    size_t len = strlen(s);

    char *t = s + len - 1;
    while (t >= s) {
        printf("%c", *t);
        t = t - 1;
    };
    puts("");
}


int main_reversor() {
    char texto[20];
    scanf("%19s", texto);
    print_reverse(texto);
    return 0;
}

// A ideia do livro era fazer palavra cruzada :p

int main_frutas() {
    char *juices[] = {
        "dragonfruit", "waterberry", "sharonfruit", "uglifruit", "rumberry", "kiwifruit", "mulberry", "strawberry", "blueberry", "blackberry", "starfruit"
    };
    char *a;

    // Horizontais
    puts(juices[6]);
    print_reverse(juices[7]);
    a = juices[2];
    juices[2] = juices[8];
    juices[8] = a;
    puts(juices[8]);
    print_reverse(juices[(18 + 7)/5]);

    // Verticais
    puts(juices[2]);
    print_reverse(juices[9]);
    juices[1] = juices[3];
    puts(juices[10]);
    print_reverse(juices[1]);

    return 0;
}

int main() {
    return main_frutas();
}
