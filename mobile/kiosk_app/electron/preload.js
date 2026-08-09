const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronKiosk', {
  getConfig: () => ipcRenderer.invoke('get-kiosk-config'),
  restartTerminal: () => ipcRenderer.send('restart-kiosk'),
});
