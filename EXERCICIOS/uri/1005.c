#include <stdio.h>
void main() {
	float x, y;
	scanf("%f %f", &x, &y);
	double media = (x * 3.5 + y * 7.5) / (double) 11;
	printf("MEDIA = %.5f\n", media);
}
