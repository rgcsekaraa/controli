package controli

import (
	"embed"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
)

//go:embed web/assets/*
var webAssets embed.FS

func webAsset(name string) []byte {
	data, _ := webAssets.ReadFile("web/assets/" + name)
	return data
}

func serveAsset(w http.ResponseWriter, contentType string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

func RenderWebTerminalHTML(token string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Controli Terminal</title>
  <link rel="stylesheet" href="/assets/xterm.css?token=%[1]s">
  <style>
    html, body { width: 100%%; height: 100%%; margin: 0; background: #0c0c0c; overflow: hidden; }
    #bar { height: 28px; display: flex; align-items: center; gap: 8px; padding: 0 8px; box-sizing: border-box; color: #d7d7d7; background: #151515; font: 12px/28px system-ui, sans-serif; }
    #status { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    #download { height: 22px; border: 1px solid #555; background: #242424; color: #f1f1f1; font: 12px system-ui, sans-serif; cursor: pointer; }
    #terminal { width: 100%%; height: calc(100%% - 28px); }
    .xterm { height: 100%%; padding: 8px; box-sizing: border-box; }
  </style>
</head>
<body>
  <div id="bar"><div id="status">connecting</div><button id="download" type="button">Download</button></div>
  <div id="terminal"></div>
  <script src="/assets/xterm.js?token=%[1]s"></script>
  <script src="/assets/xterm-addon-fit.js?token=%[1]s"></script>
  <script>
    const token = %[2]q;
    const controlPrefix = '\x00CONTROLI:';
    const status = document.getElementById('status');
    const downloadButton = document.getElementById('download');
    const term = new Terminal({
      cursorBlink: true,
      convertEol: false,
      scrollback: 10000,
      fontFamily: 'Cascadia Mono, Consolas, Menlo, monospace',
      fontSize: 14,
      theme: { background: '#0c0c0c' },
      windowsMode: true
    });
    const fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.open(document.getElementById('terminal'));
    fitAddon.fit();
    term.focus();
    const wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const clientStorageKey = 'controli.client.' + token;
    let clientId = localStorage.getItem(clientStorageKey);
    if (!clientId) {
      const random = new Uint8Array(16);
      crypto.getRandomValues(random);
      clientId = Array.from(random, b => b.toString(16).padStart(2, '0')).join('');
      localStorage.setItem(clientStorageKey, clientId);
    }
    let socket = null;
    let reconnectDelay = 250;
    let manuallyClosed = false;
    let pendingInput = [];
    const downloads = new Map();
    function socketURL() {
      return wsProtocol + '//' + location.host + '/ws?token=' + encodeURIComponent(token) + '&client_id=' + encodeURIComponent(clientId);
    }
    function flushInput() {
      if (!socket || socket.readyState !== WebSocket.OPEN) return;
      while (pendingInput.length > 0 && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: 'input', data: pendingInput.shift() }));
      }
    }
    function sendResize() {
      if (socket && socket.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'resize', columns: term.cols, rows: term.rows }));
    }
    function sendDownloadRequest() {
      const path = prompt('Download from controli-drive:');
      if (!path) return;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        status.textContent = 'download unavailable while disconnected';
        return;
      }
      const id = crypto.randomUUID ? crypto.randomUUID() : String(Date.now()) + Math.random();
      downloads.set(id, { name: 'download', chunks: [], received: 0, size: 0 });
      socket.send(JSON.stringify({ type: 'download_request', id, path }));
      status.textContent = 'download requested';
    }
    function handleControl(payload) {
      if (!payload || !payload.type) return false;
      if (payload.type === 'download_start') {
        downloads.set(payload.id, { name: payload.name || 'download', chunks: [], received: 0, size: payload.size || 0 });
        status.textContent = 'downloading ' + (payload.name || 'file');
        return true;
      }
      if (payload.type === 'download_chunk') {
        const item = downloads.get(payload.id);
        if (!item) return true;
        const raw = atob(payload.data || '');
        const bytes = new Uint8Array(raw.length);
        for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
        item.chunks.push(bytes);
        item.received += bytes.byteLength;
        status.textContent = item.size > 0 ? ('downloading ' + item.name + ' ' + Math.floor(item.received * 100 / item.size) + '%%') : ('downloading ' + item.name);
        return true;
      }
      if (payload.type === 'download_done') {
        const item = downloads.get(payload.id);
        if (!item) return true;
        const blob = new Blob(item.chunks, { type: 'application/octet-stream' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = item.name || 'download';
        document.body.appendChild(link);
        link.click();
        link.remove();
        URL.revokeObjectURL(url);
        downloads.delete(payload.id);
        status.textContent = 'download complete: ' + item.name;
        return true;
      }
      if (payload.type === 'download_error') {
        downloads.delete(payload.id);
        status.textContent = 'download failed: ' + (payload.error || 'unknown error');
        return true;
      }
      return false;
    }
    function handleSocketData(data) {
      if (data instanceof ArrayBuffer) {
        const bytes = new Uint8Array(data);
        if (bytes.length >= controlPrefix.length) {
          const prefix = new TextDecoder().decode(bytes.slice(0, controlPrefix.length));
          if (prefix === controlPrefix) {
            try {
              if (handleControl(JSON.parse(new TextDecoder().decode(bytes.slice(controlPrefix.length))))) return;
            } catch {}
          }
        }
        term.write(bytes);
        return;
      }
      if (typeof data === 'string' && data.startsWith(controlPrefix)) {
        try {
          if (handleControl(JSON.parse(data.slice(controlPrefix.length)))) return;
        } catch {}
      }
      term.write(data);
    }
    function connect() {
      socket = new WebSocket(socketURL());
      socket.binaryType = 'arraybuffer';
      socket.addEventListener('open', () => {
        status.textContent = 'connected';
        reconnectDelay = 250;
        sendResize();
        flushInput();
        term.focus();
      });
      socket.addEventListener('message', (event) => {
        handleSocketData(event.data);
      });
      socket.addEventListener('close', () => {
        if (manuallyClosed) return;
        status.textContent = 'reconnecting';
        window.setTimeout(connect, reconnectDelay);
        reconnectDelay = Math.min(reconnectDelay * 2, 5000);
      });
      socket.addEventListener('error', () => {
        if (socket) socket.close();
      });
    }
    term.onData((data) => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: 'input', data }));
        return;
      }
      pendingInput.push(data);
      while (pendingInput.length > 256) pendingInput.shift();
    });
    term.onResize(sendResize);
    downloadButton.addEventListener('click', sendDownloadRequest);
    window.addEventListener('resize', () => { fitAddon.fit(); sendResize(); });
    window.addEventListener('beforeunload', () => { manuallyClosed = true; if (socket) socket.close(); });
    connect();
  </script>
</body>
</html>`, token, token)
}

func openBrowser(target string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	case "darwin":
		command = exec.Command("open", target)
	default:
		command = exec.Command("xdg-open", target)
	}
	_ = command.Start()
}
