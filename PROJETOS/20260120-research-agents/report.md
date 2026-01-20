# Comparative Analysis: Tooling and Prompting in Code Agent Implementations

This report analyzes three prominent code agent implementations—**Claude Code**, **Crush**, and **OpenCode**—focusing on how they structure tools, design prompts, and implement best practices for accuracy and efficiency.

---

## 1. Tool Structuring and Vocabulary

The three agents take distinct approaches to tool definition, ranging from plugin-centric to documented literal constraints.

### Claude Code: MCP-First & Specialization
*   **MCP Integration:** Treats the **Model Context Protocol (MCP)** as a first-class citizen. Tools are dynamically discovered from MCP servers and prefixed (e.g., `mcp__plugin_<plugin-name>_<tool-name>`).
*   **Declarative Scope:** Uses **Markdown frontmatter** to declare `allowed-tools` for specialized agents (Explorer, Architect, Reviewer), reducing hallucinations by limiting the model's action space.
*   **Blueprint Approach:** Specialized agents (like `code-architect`) are designed to produce "actionable blueprints" before any code is modified.

### Crush: Rigorous Literal Constraints
*   **Documented Tools:** Attaches extensive **Markdown manuals** to each tool (e.g., `edit.md`). These contain "Critical Requirements," "Whitespace Checklists," and "Recovery Steps" injected directly into the LLM's context.
*   **Literal Matching:** The `edit` tool enforces **exact character matching** (spaces, tabs, newlines). The prompt includes a "whitespace checklist" to force the LLM to count spaces and verify indentation before submission.
*   **Atomic Grouping:** Uses a `Multiedit` tool to group related changes, ensuring consistency across multiple files in a single turn.

### OpenCode: Modular delegation & LSP
*   **Agent Delegation:** Uses a `Task` tool to spawn sub-agents (`general`, `explore`, `build`). Each sub-agent is filtered by a `PermissionNext` system that whitelists specific tools.
*   **Intelligence Tools:** Expõe ferramentas de **LSP (Language Server Protocol)** para fornecer diagnósticos, definições e referências de símbolos, indo além da simples leitura de arquivos.
*   **Parallel execution:** Explicitly instructs the model to execute independent tool calls in parallel (e.g., multiple `read` or `grep` calls), maximizing throughput.

---

## 2. Context Assembly Architecture

How each agent builds the "System Context" determines quão "consciente" do projeto a LLM se torna.

### Crush: Go Templates & XML Skills
*   **Dynamic Skills:** Scans the project for "Skills" (custom capabilities) and injects them as a structured **XML block** (`<skills>...</skills>`) in the system prompt.
*   **Git State Integration:** Injects a "Git Status Summary" (last 20 lines of changes) and the last 3 commit messages to provide historical context.
*   **Context Paths:** Allows defining fixed `ContextPaths` in configuration that are always read and injected into the prompt (e.g., `README.md`, `ARCHITECTURE.md`).

### OpenCode: Multi-layered Instructions & Rule Files
*   **Rule File Hierarchy:** Automatically searches for and injects instructions from `CLAUDE.md`, `AGENTS.md`, and `CONTEXT.md`. It supports both project-local and user-global rule files.
*   **Environment Block:** Injects a standardized `<env>` block with the working directory, platform, date, and a directory tree summary.
*   **Model-Specific Prompts:** Swaps the base system prompt depending on the model being used (Gemini, Anthropic, or GPT), optimizing the instructions for each provider's strengths.

### Claude Code: Discovery-Driven
*   **Pattern Discovery:** Context is built by first forcing the agent to find similar features and "technology stack boundaries."
*   **SDK Focus:** When using the Agent SDK, it injects specific verification checklists to ensure the generated code follows SDK best practices.

---

## 3. Security, Permissions, and Lifecycle Management

Safety is handled differently across the implementations, from hardcoded lists to complex hook systems.

### Crush: Safe Command Lists
*   **Hardcoded Whitelists:** Maintains a list of `safeCommands` (e.g., `git status`, `ls`, `df`). Commands outside this list might be restricted or flagged.
*   **OS-Specific Safety:** Implements specialized lists for Windows (e.g., `ipconfig`, `tasklist`) vs. Unix-like systems.

### OpenCode: PermissionNext Engine
*   **Granular Actions:** Implements a three-state permission engine: `allow`, `deny`, or `ask`.
*   **Pattern Matching:** Supports wildcard-based rules for specific tools and paths (e.g., denying `read` for `*.env` files or requiring `ask` for all destructive `bash` commands).
*   **Session Suspension:** The `ask` state suspends execution and triggers a UI event (`permission.asked`) to wait for manual user approval before resuming.

### Claude Code: Plugin Hooks
*   **Event Lifecycle:** Uses a hook system (`hookify`) with events like `PreToolUse`, `PostToolUse`, `Stop`, and `UserPromptSubmit`.
*   **Python Interceptors:** Allows developers to write Python scripts that intercept tool calls, validate arguments against custom rules, and potentially modify or block the execution.

---

## 4. System Prompt Design and Operational Guidelines

Os prompts de sistema são o "motor de comportamento" dos agentes, utilizando técnicas de condicionamento e payloads estruturados.

### Crush: O Especialista Minimalista
*   **Prompt de Tarefa e Inicialização**: O Crush foca em **respostas de uma palavra** e **autonomia absoluta**. Seus prompts (`task.md.tpl`) proíbem introduções ("Here is what I found") e conclusões, tratando a LLM como um componente de software.
*   **Payload de Contexto**: O Crush envia o estado completo do Git e XMLs de "Skills" (capacidades extras descobertas no projeto) para que o modelo saiba exatamente o que pode fazer sem perguntar ao usuário.
*   **Instruções de Edição**: O manual da ferramenta `edit.md` é injetado recursivamente, reforçando a necessidade de correspondência literal de whitespace.

### OpenCode: Psicologia de Prompt e Modos de Operação
*   **"Beast Mode" (`beast.txt`)**: Utiliza termos como "UNACCEPTABLE" e "MUST iterate" para forçar a resolução de problemas complexos. É um prompt focado em persistência e autonomia total.
*   **Guidelines de UI/UX (`codex_header.txt`)**: Diferente dos outros, o OpenCode tem diretrizes fortes contra designs "bland" (genéricos), instruindo o modelo a escolher tipografia expressiva e evitar layouts boilerplate.
*   **System Reminders**: Injeta lembretes recorrentes no histórico para evitar que o modelo esqueça mandatos críticos durante sessões longas. Exemplo:
    ```xml
    <system-reminder>
    Your operational mode has changed from plan to build.
    You are no longer in read-only mode.
    You are permitted to make file changes, run shell commands, and utilize your arsenal of tools as needed.
    </system-reminder>
    ```

### Claude Code: Arquitetura e Missões
*   **Blueprint Decisivo**: Os prompts do Claude Code exigem que o modelo produza um plano final e aja sobre ele, proibindo o modelo de apresentar múltiplas opções ao usuário.
*   **Especialização por Sub-agente**: Cada agente (Reviewer, Explorer, Architect) tem um prompt de sistema focado em métricas de sucesso diferentes (ex: fidelidade vs. cobertura de caminhos).

---

## 5. Best Practices for Accuracy and Efficiency

| Feature | Claude Code | Crush | OpenCode |
| :--- | :--- | :--- | :--- |
| **Accuracy** | Frontmatter tool-allowlists | **Literal Match Manuals** | `Read-before-Edit` mandate |
| **Efficiency** | MCP Caching | Atomic Multiedits | **Parallel Tool Execution** |
| **Safety** | Plugin Hooks | **Safe Command Whitelists** | `PermissionNext` engine |
| **Context** | Pattern Discovery phase | XML Skill injection | **Hierarchical Rule Files** |

### Key Best Practice: The "Read-Before-Write" Mandate
All three agents strictly enforce a discovery phase. **Crush** requires using the `View` tool to copy exact text before an `Edit`. **OpenCode** defines a specific "Software Engineering Workflow": Understand → Plan → Implement → Verify.

### Key Best Practice: Self-Verification and Recovery
**Crush** provides a "Recovery Steps" guide in its tool prompts, telling the agent how to handle "string not found" errors (e.g., "copy more context," "check for trailing spaces"). **OpenCode** mandates running project-specific linting/tests after every change.

---

## 6. Conclusion

*   **Crush** is optimized for **precision and fidelity** in small-to-medium edits where exact formatting is critical.
*   **Claude Code** excels at **architectural consistency** and leveraging external ecosystems via MCP and extensible hooks.
*   **OpenCode** oferece a **orquestração mais robusta**, usando sub-agentes especializados, execução paralela e um motor de permissões granular para lidar com tarefas complexas em codebases de larga escala.
