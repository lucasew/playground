# AGENTS.md - Instruções para Assistentes

## Objetivos e Restrições

1. **Objetivo Principal:** 
   - Mover projetos para dentro da pasta `PROJETOS/`
   - Manter hierarquia organizada por data

2. **O que NÃO mover:**
   - Pastas com nomes em MAIÚSCULAS (`ARQUIVO/`, `MODELOS/`, `PROJETOS/`)
   - `exercism/` (exceção especial)
   - Arquivo `be-careful` (manter na raiz)

3. **Padrão de nomenclatura de destino:**
   - Formato: `YYYYMMDD-nome-descritivo-em-kebab-case/`
   - Data do último commit (não da criação)
   - Exemplo: `20220501-fsharp-powershell-cmdlet/`

## Fluxo de Trabalho para a IA

1. **Análise:**
   ```bash
   git log -1 --format="%ad" --date=short -- [caminho]
   git log -1 --format="%s" -- [caminho]
   list_directory [caminho]
   ```

2. **Sugestão:**
   - Apresentar contexto e proposta
   - AGUARDAR APROVAÇÃO do usuário
   - NUNCA executar sem aprovação prévia

3. **Execução (após aprovação):**
   ```bash
   mkdir -p PROJETOS/YYYYMMDD-nome-projeto
   mv [caminho-original]/* PROJETOS/YYYYMMDD-nome-projeto/
   rm -rf [pasta-vazia]  # se ficou vazia
   git add .
   git commit -m "refactor: move [projeto] to PROJETOS/"
   ```

4. **Limpar após migração:**
   - Sempre remover diretórios vazios
   - Exceto se contiverem `.gitkeep`

## Regras de Commits

- **Um commit por projeto** - nunca combinar migrações
- **Fazer commit imediatamente após migração**
- **Mensagem padrão:** `refactor: move [projeto] to PROJETOS/`
- **Direto e simples** - evitar descrições longas

## Status Atual

- ✅ **Mantidos na raiz:** `ARQUIVO/`, `MODELOS/`, `PROJETOS/`, `be-careful`, `exercism/`
- ✅ **Já migrados:** `arduino/`, `assembly/`, `auxilio_pipeline/`, `blender/`, `browser/`, `elixir/`, `f-sharp/`
- 🔄 **Pendentes:** `c/`, `cgi/`, `challenges/`, `cmake/`, `docker/`, `docker-compose/`, `electron/`, `facul/`, `gemini/`, `git/`, `github/`, `golang/`, `guix/`, `help/`, `html/`, `java/`, `kivy/`, `kubernetes/`, `lsp/`, `markdown/`, `nix/`, `nodejs/`, `ocaml/`, `openziti/`, `pandas_draw/`, `pandoc/`, `php/`, `piton/`, `poc/`, `poc_c/`, `projetos/`, `protobuf/`, `revealjs/`, `rust/`, `shell/`, `sqlite/`, `talks/`, `terraform/`, `tools/`, `typst/`, `vagrant/`, `varnish/`, `vercel/`, `zig/`
