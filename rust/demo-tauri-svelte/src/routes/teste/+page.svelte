<script type="typescript">
	import { onMount } from "svelte";
    import { ask, open, message } from '@tauri-apps/api/dialog';
    import {readTextFile} from '@tauri-apps/api/fs'
	import { Button } from "sveltestrap";


    let counter = 0;
    onMount(() => {
        const interval = setInterval(() => counter++, 100)
        return () => clearInterval(interval)
    })
    let textFile = ''

    async function askTextFile() {
        const file = await open({
            title: "Manda bala",
            filters: [{
                name: 'Text',
                extensions: ['txt', 'html']
            }],
            multiple: false
        })
        if (typeof file != 'string') {
            await message("No file was supplied", {
                type: 'error'
            })
            return
        }
        textFile = await readTextFile(file)
    }

    async function askUser() {
        console.log('ask')
        console.log(await ask("Teste", 'Eoq'))
    }
</script>

<h1>Counter: {counter}</h1>

<Button on:click={askUser}>Dialog</Button>
<Button on:click={askTextFile}>Ask text file</Button>
<pre>
{textFile}
</pre>