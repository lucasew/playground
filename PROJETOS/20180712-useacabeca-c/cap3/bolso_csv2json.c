#include <stdio.h>

int main () {
    // Me chame usando pipe com aquele arquivo csv se quiser me testar, tipo arquivo.csv | programa
    // Pra que implementar esquema de arquivo se quando você passar um fluxo de csv eu já converto e gero um fluxo de json
    // cat arquivo.csv | eu > arquivo.json
    float latitude, longitude;
    char info[80];
    int linha = 0;
    int started = 0;
    puts("data=[");
    while(scanf("%f,%f,%79[^\n]", &latitude, &longitude, info) == 3) {
        if (started) printf(",\n");
        else started = 1;
        linha++;
        if (latitude < -90 || latitude > 90) {
            fprintf(stderr, "Latitude inválida em linha %i\n", linha);
            return 2;
        }
        if (longitude < -180 || longitude > 180) {
            fprintf(stderr, "Longitude inválida em linha %i\n", linha);
            return 2;
        }
        printf("{latitude: %f, longitude: %f, info: '%s'}", latitude, longitude, info);
    }
    puts("\n]");
    return 0;
}
