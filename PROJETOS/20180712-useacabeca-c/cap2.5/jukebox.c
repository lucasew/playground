#include <stdio.h>
#include <string.h>

char tracks[][80] = {
 "I left my heart in Harvard Med School",
 "Newark, Newark - a wonderful town",
 "Dancing with a Dork",
 "From here to maternity",
 "The girl from Iwo Jima",
};

void find_track(char query[])  {
    int i;
    for (i = 0; i < 5; i++) {
        if (strstr(tracks[i], query)) {
            //printf("Achado id:%i: %s\n", i, tracks[i]);
            puts(tracks[i]);
        }
    }
}

int main () {
    char query[80];
    printf("Pesquisar por?: ");
    scanf("%80s", query);
    //fgets(query, 80, stdin); // apenas com scanf funciona :/
    find_track(query);
    return 0;
}
