#include <stdio.h>
void main() {
	double in;
	scanf("%lf", &in);
	double res = (double) 3.14159 * in * in;
	printf("A=%.4f\n", res);
}
