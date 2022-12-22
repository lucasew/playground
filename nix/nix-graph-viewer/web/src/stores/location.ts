import { derived, writable } from "svelte/store";

export function createUrlStore() {
    const hash = writable(new URL(window.location.href).hash)

    const originalPushState = history.pushState;
    const originalReplaceState = history.replaceState;

    const updateHref = () => {
        let newHash = new URL(window.location.href).hash
        if (newHash.length > 0 && newHash[0] == "#") {
            newHash = newHash.slice(1)
        }
        return hash.set(newHash)
    };

    hash.subscribe(h => console.log('hash', h))

    history.pushState = function (...args) {
        originalPushState.apply(this, args);
        updateHref();
    };

    history.replaceState = function (...args) {
        originalReplaceState.apply(this, args);
        updateHref();
    };

    window.addEventListener("popstate", updateHref);
    window.addEventListener("hashchange", updateHref);
    return hash
}
export function handleJump(route: string) {
    let url = new URL(window.location.href)
    url.hash = route
    return () => history.pushState({}, '', url.toString())
}

export const locationStore = createUrlStore();

history.replaceState({}, '', window.location.href)

