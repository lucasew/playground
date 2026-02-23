# Objetivo
Você está trabalhando no projeto `go-csmith` para obter equivalência real com o Csmith upstream (C++).

## Regra principal
Use o relatório de divergência RNG como verdade operacional: corrija o code path em Go para empurrar o primeiro desvio para frente, sem hacks.

## O que fazer em cada iteração
1. Ler o desvio atual (evento, contexto upstream/go).
2. Identificar a função/caminho no Go responsável pelo desvio.
3. Portar comportamento upstream de forma fiel (ordem de chamadas RNG, filtros, retries, profundidade/contexto).
4. Evitar mudanças cosméticas e evitar alterações no algoritmo upstream.
5. Fazer mudanças mínimas e coerentes para avançar a paridade.

## Restrições
- Não trapacear.
- Não mascarar resultado removendo trechos de geração.
- Não “alinhar por sorte” com consumos artificiais sem base no upstream.
- Priorizar fidelidade de pipeline e RNG call path.

## Definição de pronto
- O primeiro desvio de RNG deve avançar para um evento maior a cada iteração, idealmente até `result=match`.
- A implementação Go deve continuar compilando e gerando código válido.

## Saída esperada da iteração
- Resumo curto:
  - qual era o desvio
  - qual mudança foi aplicada
  - novo primeiro desvio
  - próximos passos

