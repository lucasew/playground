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

### Regra de Commit Obrigatório

- **SEMPRE fazer commit após cada migração de projeto**
- **NÃO tentar emendar a próxima fase antes da anterior estar commitada**
- **Cada projeto movido deve ter seu próprio commit**
- **Usar mensagens descritivas**: `refactor: move [projeto] to PROJETOS/`

### Mensagens de Commit

- **Manter mensagens simples e diretas**
- **Não complicar demais as descrições**
- **Formato básico**: `refactor: move [projeto] to PROJETOS/`
- **Adicionar linha adicional apenas se necessário** para contexto específico
- **Evitar textos longos e desnecessários**

### Enriquecimento de Contexto

- **SEMPRE capturar a data do último commit** usando `git log -1 --format="%ad" --date=short -- [arquivo]`
- **SEMPRE capturar a mensagem do último commit** usando `git log -1 --format="%s" -- [arquivo]`
- **Usar essas informações para enriquecer o contexto** na sugestão do projeto
- **Ajuda a entender o propósito e histórico** de cada projeto

### Avaliação Crítica de Projetos

- **Ser crítico e honesto** sobre a qualidade/utilidade dos projetos
- **Não "puxar saco"** - avaliar objetivamente
- **Identificar projetos que podem ser removidos** ao invés de movidos
- **Questionar se vale a pena manter** projetos antigos/obsoletos
- **Sugerir remoção** quando apropriado

### Fluxo de Trabalho

1. **Atualizar AGENTS.md** com novas instruções se necessário
2. **Fazer commit do AGENTS.md** antes de sugerir próximo projeto
3. **Capturar data e mensagem do último commit** do projeto
4. **Sugerir projeto** com contexto enriquecido e avaliação crítica
5. **Executar migração** após aprovação
6. **Fazer commit da migração**
7. **Repetir processo**

### Pastas que devem permanecer na raiz (nomes maiúsculos):
- `ARQUIVO/`
- `MODELOS/`
- `PROJETOS/`
- `be-careful` (arquivo especial - manter na raiz)
- Outras pastas com nomes em maiúsculo

### Pastas que devem ser migradas para `PROJETOS/`:
- `arduino/` ✅ (removido - sketch padrão)
- `assembly/` ✅ (movido para 20230530-assembly-print-vector)
- `auxilio_pipeline/` ✅ (movido para 20201114-auxilio-emergencial-pipeline)
- `be-careful` ✅ (mantido na raiz - arquivo especial)
- `blender/` ✅ (movido para 4 projetos separados)
- `browser/` ✅ (movido para 20210305-golang-chromedp-gdocs-typing)
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
