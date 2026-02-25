#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main() {
    char line[80];
    FILE *in = fopen("spooky.csv", "r");
    FILE *file1 = fopen("ufos.csv", "w");
    FILE *file2 = fopen("disappearances.csv", "w");
    FILE *file3 = fopen("others.csv", "w");

    while (scanf(in, "%79[^\n]", line) == 1) {
        if (strstr(line, "UFO"))
            fprintf(file1, "%s\n", line);
        else if (strstr(line, "Disappearance") == 1)
            fprintf(file2, "%s\n", line);
        else
            fprintf(file3, "%s\n", line);
    }
    fclose(file1);
    fclose(file2);
    fclose(file3);
    return 0;
}
