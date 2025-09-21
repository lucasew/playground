<script lang="ts">
    import 'reveal.js/dist/reveal.css';
    import 'reveal.js/dist/reset.css';
    import 'reveal.js/dist/theme/black.css';
    import 'reveal.js/plugin/highlight/monokai.css';

    import Highlight from 'reveal.js/plugin/highlight/highlight';
    import Markdown from 'reveal.js/plugin/markdown/markdown';
    import MathReveal from 'reveal.js/plugin/math/math';
    import RevealNotes from 'reveal.js/plugin/notes/notes';
    import RevealSearch from 'reveal.js/plugin/search/search';

    import Reveal from 'reveal.js';

    import { onMount, tick } from 'svelte';

    onMount(async () => {
        console.log("reveal")
        const revealConfig = {
            controls: true,
            progress: true,
            center: true,
            hash: true,
            plugins: [
                Highlight,
                Markdown,
                MathReveal,
                RevealNotes,
                RevealSearch
            ]
        }
        await tick();
        const deck = new Reveal(revealConfig);
        deck.initialize();
      });

</script>

<div class="reveal">
    <div class="slides">
        <slot/>
    </div>
</div>
