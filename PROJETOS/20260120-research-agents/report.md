# Comparative Analysis: Tooling and Prompting in Code Agent Implementations

This report analyzes three prominent code agent implementations—**Claude Code**, **Crush**, and **OpenCode**—focusing on how they structure tools, design prompts, and implement best practices for accuracy and efficiency.

---

## 1. Tool Structuring and Integration

The three agents take distinct approaches to tool definition and registration, ranging from plugin-centric to tightly integrated language-specific implementations.

### Claude Code: Plugin-Centric & MCP-First
*   **MCP Integration:** Claude Code treats the **Model Context Protocol (MCP)** as a first-class citizen. Tools are often dynamically discovered from MCP servers and prefixed (e.g., `mcp__plugin_<plugin-name>_<tool-name>`).
*   **Declarative Tooling:** Commands and specialized agents use **Markdown frontmatter** to declare `allowed-tools`. This limits the agent's scope to only what is necessary for the task, improving efficiency and reducing hallucinations.
*   **Schema-Driven:** Relies heavily on `inputSchema` (JSON Schema) to define tool parameters, ensuring the LLM provides correctly typed inputs.

### Crush: Go-Based & Documented Tools
*   **Embedded Documentation:** Crush implements tools as Go structs but attaches extensive **Markdown documentation** to each (e.g., `edit.md`). These files contain "Critical Requirements," "Whitespace Checklists," and "Recovery Steps" that are injected into the context.
*   **Strict Constraints:** Tools like `edit` have extremely literal matching requirements (exact whitespace/indentation), which are enforced by the tool logic and emphasized in the documentation.
*   **Dynamic Assembly:** Uses Go templates to assemble prompts and tool definitions dynamically based on the current session's "skills."

### OpenCode: Monorepo & Parallel Execution
*   **Centralized Functions:** Tools are organized within a TypeScript monorepo (`packages/function`).
*   **Parallelism Focus:** The system prompt explicitly instructs the agent to "Execute multiple independent tool calls in parallel when feasible," optimizing performance for search and read operations.
*   **Path Enforcement:** Implements a "Core Mandate" for **Absolute Path Construction**, requiring the agent to resolve all relative paths to the project root before calling any filesystem tool.

---

## 2. Prompt Engineering Strategies

Prompts are the "brain" of these agents. The implementations vary from specialized multi-agent systems to comprehensive single-prompt instructions.

### Claude Code: Multi-Agent Specialization
*   **Specialized Roles:** Instead of one large prompt, it uses several specialized agents (Explorer, Architect, Reviewer) each with its own focused system prompt (e.g., `code-explorer.md`).
*   **Mission-Driven:** Prompts are structured around a "Core Mission" and specific "Analysis Approaches," providing a clear methodology for the agent to follow.

### Crush: Template-Driven & XML Structuring
*   **Contextual Injection:** Uses Go templates to inject discovered "Skills" as XML into the system prompt.
*   **Instructional XML:** Uses tags like `<prerequisites>`, `<critical_requirements>`, and `<best_practices>` within tool documentation to help the LLM parse complex rules more effectively.

### OpenCode: Comprehensive Mandates & Workflow
*   **Core Mandates:** Uses a very detailed system prompt organized into high-level sections: Core Mandates, Primary Workflows, and Operational Guidelines.
*   **Chain-of-Thought Examples:** Includes extensive examples of how the agent should think ("First, I'll analyze...", "Great, tests exist...") and call tools sequentially to solve complex refactoring or testing tasks.
*   **Tone Control:** Explicitly mandates a "Concise & Direct" tone with "Minimal Output" (fewer than 3 lines of text) to maintain focus on the CLI experience.

---

## 3. Best Practices for Accuracy and Efficiency

Each implementation incorporates specific guardrails to ensure reliability and speed.

| Feature | Claude Code | Crush | OpenCode |
| :--- | :--- | :--- | :--- |
| **Accuracy** | Tool pre-allowing (least privilege) | Strict whitespace matching & checklists | "Read before Write" & self-verification loops |
| **Efficiency** | Cached MCP results & batching | Dynamic skill injection | Parallel tool execution |
| **Safety** | User confirmation for destructive tools | Recovery steps for failed edits | Absolute path construction mandate |
| **Validation** | Schema-based parameter validation | Pre-edit verification (LS/View) | Build & Lint checks after edits |

### Key Best Practice: The "Read-Before-Write" Loop
All three agents emphasize a discovery phase. **Crush** mandates using the `View` tool to copy exact text before an `Edit`. **OpenCode** defines a "Software Engineering Task" sequence: Understand (Grep/Glob) → Plan → Implement → Verify.

### Key Best Practice: Self-Correction
**Crush** provides a "Recovery Steps" guide in its tool prompts, telling the agent exactly what to do if an edit fails (e.g., "View the file again," "Copy more context"). **OpenCode** encourages using unit tests and build commands as a feedback loop.

---

## 4. Conclusion

*   **For Accuracy:** **Crush's** approach of providing literal, documented constraints for every tool (like a "whitespace checklist") is highly effective for reducing file-edit failures.
*   **For Complex Tasks:** **Claude Code's** multi-agent approach allows for deep specialization, making it better at high-level architecture or thorough code reviews.
*   **For Performance:** **OpenCode's** emphasis on parallel tool execution and minimal conversational overhead makes it optimized for fast, iterative CLI usage.
