const {contextBridge, ipcRenderer} = require('electron')

function handleTypeSearch(query) {
    return ipcRenderer.invoke('searchInput', query)
}
function setExpanded(isExpanded = false) {
    return ipcRenderer.invoke('setExpanded', isExpanded)
}
contextBridge.exposeInMainWorld('app', {
    handleTypeSearch,
    setExpanded
})
