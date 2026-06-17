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
    #status { height: 24px; line-height: 24px; padding: 0 8px; box-sizing: border-box; color: #d7d7d7; background: #151515; font: 12px/24px system-ui, sans-serif; }
    #terminal { width: 100%%; height: calc(100%% - 24px); }
    .xterm { height: 100%%; padding: 8px; box-sizing: border-box; }
  </style>
</head>
<body>
  <div id="status">connecting</div>
  <div id="terminal"></div>
  <script src="/assets/xterm.js?token=%[1]s"></script>
  <script src="/assets/xterm-addon-fit.js?token=%[1]s"></script>
  <script>
    const token = %[2]q;
    const status = document.getElementById('status');
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
    const socket = new WebSocket('ws://' + location.host + '/ws?token=' + encodeURIComponent(token));
    socket.binaryType = 'arraybuffer';
    function sendResize() {
      if (socket.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'resize', columns: term.cols, rows: term.rows }));
    }
    socket.addEventListener('open', () => { status.textContent = 'connected'; sendResize(); term.focus(); });
    socket.addEventListener('message', (event) => {
      if (event.data instanceof ArrayBuffer) term.write(new Uint8Array(event.data));
      else term.write(event.data);
    });
    socket.addEventListener('close', () => { status.textContent = 'closed'; term.write('\r\n\x1b[31mconnection closed\x1b[0m\r\n'); });
    term.onData((data) => {
      if (socket.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'input', data }));
    });
    term.onResize(sendResize);
    window.addEventListener('resize', () => { fitAddon.fit(); sendResize(); });
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
