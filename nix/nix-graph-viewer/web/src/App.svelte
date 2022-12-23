<script lang="ts">
import {onMount} from 'svelte'
import {locationStore, handleJump} from './stores/location.ts'
import {dataStore} from './stores/data.ts'

import ComponentTable from './lib/ComponentTable.svelte'

onMount(() => {
    import('/tmp/cool_graph.json').then(d => {
        window.setData(d)
    })
})

function getLinksOf(data: any, id: string, is_backlink: boolean = false) {
    const key = id.replace('nix:', '/nix/store/')
    return ((is_backlink ? data.backlinks : data.links)[key] || []).filter(x => x !== key)
}

/* $: console.log('p', Object.keys($dataStore.paths)) */

</script>

<h1 style='cursor: pointer' on:click={handleJump('')}>Nix closure viewer</h1>
{#if $locationStore == ""}
    {#if $dataStore.paths !== undefined}
        <ComponentTable/>
    {/if}
{:else if $locationStore.startsWith("nix:")}
    <h1>{$locationStore.replace('nix:', '/nix/store/')}</h1>
    {#if $dataStore.links !== undefined}
        <h1>Links</h1>
        <ComponentTable paths={getLinksOf($dataStore, $locationStore, false)} />
    {/if}
    {#if $dataStore.backlinks !== undefined}
        <h1>Backlinks</h1>
        <ComponentTable paths={getLinksOf($dataStore, $locationStore, true)} />
    {/if}
{/if}

