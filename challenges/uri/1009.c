#include <stdio.h>
void main () {
	char nome[20];
	double salario, vendas;
	scanf("%s %lf %lf", nome, &salario, &vendas);
	printf("TOTAL = R$ %.2lf\n", salario + vendas * 0.15);
}
