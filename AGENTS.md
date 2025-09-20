# AGENTS.md - Instruções para Assistentes

## Regras Básicas

1. **Objetivo Principal**: Mover projetos para pasta `PROJETOS/`
2. **Não Mover**: Pastas com nome MAIÚSCULO (`ARQUIVO/`, `MODELOS/`, `PROJETOS/`)
3. **Exceções Adicionais**: 
   - `exercism/` - manter como está
   - `be-careful` - manter na raiz
4. **Fluxo de Trabalho**:
   - Sugerir projeto → Aguardar aprovação → Executar → Commit

## Padrões de Nomenclatura

### Formato de Projetos Migrados
```
YYYYMMDD-descrição-do-projeto/
```

- **Data**: Último commit (usar `git log -1 --format="%ad" --date=short -- [path]`)
- **Descrição**: Nome kebab-case descritivo
- **Exemplo**: `20221201-arduino-bluetooth-hid/`

## Regras de Commit

- **Um commit por projeto** - nunca agrupar múltiplas migrações
- **Fazer commit imediatamente após migração**
- **Mensagem padrão**: `refactor: move [projeto] to PROJETOS/`
- **Simples e direto** - evitar textos longos

## Processo de Migração

### 1. Analisar Projeto
```bash
# Obter data e mensagem do último commit
git log -1 --format="%ad" --date=short -- [caminho]
git log -1 --format="%s" -- [caminho]

# Verificar conteúdo do projeto
list_directory [caminho]
```

### 2. Avaliar Criticamente
- Ser honesto sobre qualidade/utilidade
- Sugerir remover projetos obsoletos
- Não exagerar no valor dos projetos

### 3. Sugerir Migração
- Apresentar contexto e data
- **AGUARDAR APROVAÇÃO** do usuário
- **Nunca executar sem aprovação prévia**

### 4. Executar Migração
```bash
# Criar diretório de destino
mkdir -p PROJETOS/YYYYMMDD-nome-projeto

# Mover arquivos
mv [caminho-original]/* PROJETOS/YYYYMMDD-nome-projeto/

# Commit
git add .
git commit -m "refactor: move [projeto] to PROJETOS/"
```

### 5. Remover Pastas Vazias
- Sempre deletar diretórios vazios após migração
- Exceção: pastas com `.gitkeep`

### Status das Pastas

#### Não Mover (Manter na Raiz):
- `ARQUIVO/` - pasta maiúscula
- `MODELOS/` - pasta maiúscula  
- `PROJETOS/` - pasta maiúscula
- `be-careful` - arquivo especial
- `exercism/` - exceção solicitada

#### Já Migradas:
- `arduino/` ✅ (removido - sketch padrão)
- `assembly/` ✅ (20230530-assembly-print-vector)
- `auxilio_pipeline/` ✅ (20201114-auxilio-emergencial-pipeline)
- `blender/` ✅ (4 projetos separados)
- `browser/` ✅ (20210305-golang-chromedp-gdocs-typing)
- `elixir/` ✅ (3 projetos separados)

#### Pendentes de Migração:
- `c/`
- `cgi/` 
- `challenges/`
- `cmake/`
- `docker/`
- `docker-compose/`
- `electron/`
- `f-sharp/`
- `facul/`
- `gemini/`
- `git/`
- `github/`
- `golang/`
- `guix/`
- `help/`
- `html/`
- `java/`
- `kivy/`
- `kubernetes/`
- `lsp/`
- `markdown/`
- `nix/`
- `nodejs/`
- `ocaml/`
- `openziti/`
- `pandas_draw/`
- `pandoc/`
- `php/`
- `piton/`
- `poc/`
- `poc_c/`
- `projetos/`
- `protobuf/`
- `revealjs/`
- `rust/`
- `shell/`
- `sqlite/`
- `talks/`
- `terraform/`
- `tools/`
- `typst/`
- `vagrant/`
- `varnish/`
- `vercel/`
- `zig/`
