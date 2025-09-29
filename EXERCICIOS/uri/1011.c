#include <stdio.h>
#define PI 3.14159
void main() {
	double raio;
	scanf("%lf", &raio);
	double volume = (4.0/3.0)*PI*raio*raio*raio;
	printf("VOLUME = %.3lf\n", volume);
}
