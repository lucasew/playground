# AGENTS.md - Instruções para Assistentes

## Objetivos e Restrições

1. **Objetivo Principal:** 
   - Mover projetos para dentro da pasta `PROJETOS/`
   - Manter hierarquia organizada por data

2. **O que NÃO mover:**
   - Pastas com nomes em MAIÚSCULAS (`ARQUIVO/`, `MODELOS/`, `PROJETOS/`, `FACUL/`, `CHALLENGES/`, `EXERCISM/`)
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

5. **Após commit de migração:**
   - Se o passo de migração já estiver commitado, pode automaticamente procurar o próximo passo de migração sem aguardar nova ordem do usuário.

## Regras de Commits

- **Um commit por projeto** - nunca combinar migrações
- **Fazer commit imediatamente após migração**
- **Mensagem padrão:** `refactor: move [projeto] to PROJETOS/`
- **Direto e simples** - evitar descrições longas


