#include <stdio.h>
void main() {
	float total = 0;
	int i;
	for (i = 0; i < 2; i++) {
		int code, units;
		float price;
		scanf("%i %i %f", &code, &units, &price);
		total = total + units * price;
	}
	printf("VALOR A PAGAR: R$ %.2f\n", total);
}
