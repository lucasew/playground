import { writable } from 'svelte/store'

export const dataStore = writable({})

window.setData = (newData) => {
    dataStore.set(newData)
}
