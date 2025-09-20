#include <stdio.h>
// numero, horas no mes, ganho por hora
void main() {
	int id, horas;
	float precohora;
	scanf("%i %i %f", &id, &horas, &precohora);
	printf("NUMBER = %i\n", id);
	printf("SALARY = U$ %.2f\n", horas * precohora);
}
