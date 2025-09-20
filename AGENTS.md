# AGENTS.md - Instruções para Assistentes

## Organização

### Regras para Reorganização do Repositório

1. **Objetivo**: Concentrar todos os projetos na pasta `PROJETOS`
2. **Restrição**: NÃO mover nenhuma pasta com nome maiúsculo (ex: `ARQUIVO`, `MODELOS`, `PROJETOS`)
3. **Processo**: Migração projeto por projeto no bate-papo
   - Assistente sugere projeto para mover
   - Usuário analisa e aprova
   - Assistente executa a alteração
4. **Estrutura atual**: Projetos estão espalhados por pastas de tecnologia (ex: `golang/`, `rust/`, `python/`, etc.)
5. **Estrutura desejada**: Todos os projetos organizados dentro de `PROJETOS/`

### Padrão de Nomenclatura para Projetos

- **Formato**: `YYYYMMDD-descrição-do-projeto/`
- **Data**: Usar a data do último commit do arquivo/projeto (não data de criação)
- **Descrição**: Nome descritivo em kebab-case que identifica o projeto
- **Exemplo**: `20221201-arduino-bluetooth-hid/`

### Pastas que devem permanecer na raiz (nomes maiúsculos):
- `ARQUIVO/`
- `MODELOS/`
- `PROJETOS/`
- Outras pastas com nomes em maiúsculo

### Pastas que devem ser migradas para `PROJETOS/`:
- `arduino/`
- `assembly/`
- `auxilio_pipeline/`
- `be-careful`
- `blender/`
- `browser/`
- `c/`
- `cgi/`
- `challenges/`
- `cmake/`
- `docker/`
- `docker-compose/`
- `electron/`
- `elixir/`
- `exercism/`
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
