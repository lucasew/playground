<script lang="ts">
    import 'reveal.js/dist/reveal.css';
    import 'reveal.js/dist/reset.css';
    import Reveal            from 'reveal.js';

    import { onMount, tick } from 'svelte';

    onMount(async () => {
        console.log("reveal")
        const revealConfig = {
            controls: true,
            progress: true,
            center: true,
            hash: true,
            plugins: []
        }
        await tick();
        const deck = new Reveal(reveal);
        deck.initialize();
      });

</script>

<div class="reveal">
    <div class="slides">
        <slot/>
    </div>
</div>
