# GitHub Actions Workflow Input Modifier

Script para converter inputs de `workflow_dispatch` do tipo `string` para tipo `choice` com opções `[patch, minor, major]` em workflows do GitHub Actions.

## Motivação

Quando criamos workflows do GitHub Actions com `workflow_dispatch`, inputs do tipo `string` aceitam valores livres. O tipo `choice` restringe as opções e exibe um dropdown na interface do GitHub, melhorando UX e prevenindo erros.

## Estrutura

```
.
├── convert_to_choice.py      # Script principal
├── autorelease.yml           # Workflow de exemplo
├── real-autorelease.yaml     # Workflow real do ritm_annotation
├── test.sh                   # Script de testes
└── README.md                 # Este arquivo
```

## Requisitos

- Python 3.6+
- uv (ou Python com ruamel.yaml instalado)

## Como Usar

### Uso Básico (Inplace)

Por padrão, o script sobrescreve o arquivo original:

```bash
./convert_to_choice.py workflow.yml
```

### Salvar em Arquivo Diferente

```bash
./convert_to_choice.py workflow.yml -o workflow-modified.yml
```

### Modo Verbose

```bash
./convert_to_choice.py workflow.yml -v
```

## Exemplo de Conversão

### Antes:

```yaml
on:
  workflow_dispatch:
    inputs:
      new_version:
        description: 'New tag version'
        default: 'patch'
```

### Depois:

```yaml
on:
  workflow_dispatch:
    inputs:
      new_version:
        description: 'New tag version'
        default: 'patch'
        type: choice
        options:
        - patch
        - minor
        - major
```

## Comportamento

- Converte todos os inputs do tipo `string` (ou sem tipo especificado) para `choice`
- Adiciona opções `[patch, minor, major]`
- Preserva formatação e ordem original do YAML (usa ruamel.yaml)
- Mantém inputs de outros tipos inalterados
- Se o default não está nas opções, ajusta para `patch`

## Testando

```bash
# Executar testes
./test.sh

# Testar com workflow real
./convert_to_choice.py real-autorelease.yaml -o teste.yaml -v
```

## Casos de Uso

- Padronizar releases com choices fixos (patch/minor/major)
- Prevenir erros de digitação em versões
- Melhorar UX do workflow dispatch

## Implementação

O script usa `ruamel.yaml` para preservar formatação e comentários do YAML original, fazendo apenas as modificações necessárias nos inputs.
