const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('path');
const fs = require('fs');

let mainWindow = null;

function loadConfig() {
  const configPath = path.join(__dirname, 'kiosk_config.json');
  if (fs.existsSync(configPath)) {
    try {
      return JSON.parse(fs.readFileSync(configPath, 'utf8'));
    } catch (e) {
      console.error('Failed to parse kiosk_config.json:', e);
    }
  }
  return {
    store_id: 'STORE-BLR-001',
    device_id: 'KIOSK-TERM-DEFAULT',
    api_base_url: 'https://api.zippyra.com',
  };
}

function createWindow() {
  const config = loadConfig();

  mainWindow = new BrowserWindow({
    width: 1920,
    height: 1080,
    fullscreen: true,
    kiosk: true, // Lock terminal to single unattended app window
    autoHideMenuBar: true,
    frame: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
    },
  });

  // Serve static Flutter Web build output
  const webBuildIndex = path.join(__dirname, '../build/web/index.html');
  if (fs.existsSync(webBuildIndex)) {
    mainWindow.loadFile(webBuildIndex);
  } else {
    // Fallback for dev mode
    mainWindow.loadURL('http://localhost:8080');
  }

  // Crash Recovery Hook: Auto-reload renderer process on crash
  mainWindow.webContents.on('render-process-gone', (event, details) => {
    console.error(`[KIOSK CRASH RECOVERY] Renderer process gone (${details.reason}). Restarting renderer...`);
    setTimeout(() => {
      if (mainWindow) mainWindow.reload();
    }, 1000);
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

ipcMain.handle('get-kiosk-config', () => {
  return loadConfig();
});

ipcMain.on('restart-kiosk', () => {
  if (mainWindow) {
    mainWindow.reload();
  }
});

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});
