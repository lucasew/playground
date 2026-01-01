# Tentativa de rodar o backend do SvelteKit em Golang

- v8go: Interpretador JS
- Maior problema: servidor espera um monte de coisa no escopo do runtime, que não tá disponível. Um tanto inviável no momento.
- Macete: para unificar o index.js que o vite gerou do server (bundle.js) só rodei um `bunx esbuild --bundle main.js > bundle.js`.
