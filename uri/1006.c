#include <stdio.h>
void main() {
	float x, y, z;
	scanf("%f %f %f", &x, &y, &z);
	float media = (x*2 + y*3 + z*5) /10;
	printf("MEDIA = %.1f\n", media);
}
