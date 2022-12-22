<script lang="ts">
  import { writable, derived } from 'svelte/store'
  import { dataStore } from '../stores/data.ts'
  import { handleJump } from '../stores/location.ts'
  import {
    type ColumnDef,
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    type TableOptions
  } from '@tanstack/svelte-table'

  export interface Path {
    path: string
    closure: number
    nar: number
  }

  export let paths: string[] | undefined = undefined
  let defaultData = []

  $: {
    if ($dataStore.paths === undefined) {
        defaultData = []
    } else {
        const pathList = paths || Object.keys($dataStore.paths)
        console.log('paths', pathList)
        defaultData = pathList.map(path => {
            return {...$dataStore.paths[path], path} as Path
        }).sort((a, b) => a.closure < b.closure)
    }
  }
  $: console.log('defaultData', defaultData)
  

  const defaultColumns: ColumnDef<Path>[] = [
    {
        accessorKey: 'path',
        cell: item => item.getValue(),
        header: () => "Nix path"
    },
    {
        accessorKey: 'closure',
        cell: item => item.getValue(),
        header: () => "Closure size"
    },
    {
        accessorKey: 'nar',
        cell: item => item.getValue(),
        header: () => "NAR size"
    }
  ]

  let columnOrder = []

  const options = writable<TableOptions<Paths>>({
    data: defaultData,
    columns: defaultColumns,
    getCoreRowModel: getCoreRowModel(),
  })

  const rerender = () => {
    options.update(options => ({
      ...options,
      data: defaultData,
    }))
  }

 const setColumnOrder = updater => {
    if (updater instanceof Function) {
      columnOrder = updater(columnOrder)
    } else {
      columnOrder = updater
    }
    options.update(old => ({
      ...old,
      state: {
        ...old.state,
        columnOrder,
      },
    }))
  }

  $: rerender(defaultData)

  const table = createSvelteTable(options)

</script>

<table>
    <thead>
        {#each $table.getHeaderGroups() as headerGroup}
            <tr>
                {#each headerGroup.headers as header}
                    <th>
                        {#if !header.isPlaceholder}
                            <svelte:component this={flexRender(
                                header.column.columnDef.header,
                                header.getContext()
                            )} />
                        {/if}
                    </th>
                {/each}
            </tr>
        {/each}
    </thead>
   <tbody>
      {#each $table.getRowModel().rows as row}
        <tr style='cursor: pointer' on:click={handleJump(row._valuesCache.path.replace('/nix/store/', 'nix:'))}>
          {#each row.getVisibleCells() as cell}
            <td>
              <svelte:component
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
    <tfoot>
      {#each $table.getFooterGroups() as footerGroup}
        <tr>
          {#each footerGroup.headers as header}
            <th>
              {#if !header.isPlaceholder}
                <svelte:component
                  this={flexRender(
                    header.column.columnDef.footer,
                    header.getContext()
                  )}
                />
              {/if}
            </th>
          {/each}
        </tr>
      {/each}
    </tfoot>
</table>
