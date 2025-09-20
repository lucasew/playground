const electron = require('electron');

const {BrowserWindow, app} = electron;

app.on('ready', () => {
    const screen = electron.screen.getPrimaryDisplay();
    const { width, height } = screen.bounds
    let browser = new BrowserWindow({
        webPreferences: {
            preload: `${__dirname}/preload.js`
        },
        useContentSize: true,
        transparent: true,
        resizable: true,
        hasShadow: false,
        opacity: 0.5,
        maxWidth: 300,
        maxHeight: 300,
        alwaysOnTop: true,
        // titleBarStyle: 'hide',
        // skipTaskbar: true,
        // ignoreMouseEvents: true,
        // frame: false,
    });
    browser.setAutoHideMenuBar(true);
    browser.loadURL(`file://${__dirname}/hello.html`);
    browser.center();
    browser.on('resize', () => {
        const bounds = browser.getBounds()
        browser.setBounds({
            x: Math.floor((bounds.x + width) / 4),
            y: Math.floor((bounds.y + height) / 4)
        })
    })
    electron.ipcMain.handle('searchInput', (ev, arg) => {
        console.log("arg", arg)
        return arg.split('').map(text => {return {text}})
        return [
            {text: "hello, world"},
            {text: "eoq, trabson"},
        ]
    })
    electron.ipcMain.handle('setExpanded', (ev, ifExpanded) => {
        if (ifExpanded) {
            browser.setBounds({height: 300})
        } else {
            browser.setBounds({height: 100})
        }
    })
    browser.show();
})
