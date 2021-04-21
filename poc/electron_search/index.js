const {ipcMain, app, BrowserWindow} = require('electron');

const baseHeight = 100;
const baseWidth = 1000;
const itemsHeight = 500;

(async function() {
    await app.whenReady()
    const w = new BrowserWindow({
      width: baseWidth,
      height: baseHeight,
      minHeight: baseHeight,
      webPreferences: {
          nodeIntegration: true
      },
      titleBarStyle: 'hide',
      transparent: true,
      frame: false,
      resizable: false,
      backgroundColor: '#00ffffff',
      hasShadow: false,
    });
    w.loadURL(`file://${__dirname}/index.html`)
    w.setAlwaysOnTop(true, 'floating')
    w.setAutoHideMenuBar(true)
    w.show()
    function handleInput(text) {
        if (text.length == 0) {
            console.log("menor")
            w.setMinimumSize(baseWidth, baseHeight)
            w.setSize(baseWidth, baseHeight, true)
            w.center()
        } else {
            console.log("maior")
            w.setSize(baseWidth, itemsHeight, true)
            w.center()
        }
    }
    ipcMain.on('change-text', (event, arg) => {
        console.log(arg)
        handleInput(arg)
    })
})()
