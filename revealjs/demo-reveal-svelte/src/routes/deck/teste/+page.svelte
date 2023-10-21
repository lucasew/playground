<script lang="ts">
    import Slide from '$lib/Slide.svelte';
    import Code from '$lib/Code.svelte';
    console.log("aaaa")
</script>

<Slide>
    <h1>Teste</h1>
</Slide>
<Slide>
    <h1>Bla bla bla</h1>
    <Code trim=true lineNumbers=true>
        console.log("hello")
    </Code>
</Slide>
