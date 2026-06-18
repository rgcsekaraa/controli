package controli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

func RunConsoleRelayClient(relayURL, sessionID, secret string) int {
	relay := NewRelayClient(relayURL, sessionID, secret)
	fmt.Fprintln(os.Stderr, "connecting to remote CLI...")
	if err := relay.Connect(SideClient); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Fprintln(os.Stderr, "connected successfully. Starting remote CLI session.")
	stop := make(chan struct{})
	if peer, err := relay.peer(SideClient); err == nil {
		go peer.KeepAlive(stop)
	}
	go func() {
		_, _ = io.Copy(relayWriter{relay: relay}, os.Stdin)
		closeOnce(stop)
	}()
	for {
		select {
		case <-stop:
			relay.Close(SideClient)
			return 0
		default:
		}
		data, err := relay.Read(SideClient)
		if err != nil {
			closeOnce(stop)
			return 0
		}
		_, _ = os.Stdout.Write(data)
	}
}

type relayWriter struct {
	relay *RelayClient
}

func (w relayWriter) Write(p []byte) (int, error) {
	if err := w.relay.Send(SideClient, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

type WebTerminalBridge struct {
	relay        *RelayClient
	token        string
	stop         chan struct{}
	clients      map[*websocket.Conn]*sync.Mutex
	pending      [][]byte
	pendingBytes int
	mu           sync.Mutex
	upgrader     websocket.Upgrader
}

const maxWebPendingBytes = 16 * 1024 * 1024

func NewWebTerminalBridge(relay *RelayClient, token string) *WebTerminalBridge {
	return &WebTerminalBridge{
		relay:   relay,
		token:   token,
		stop:    make(chan struct{}),
		clients: map[*websocket.Conn]*sync.Mutex{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				host, _, err := net.SplitHostPort(r.Host)
				if err != nil {
					host = r.Host
				}
				return host == "127.0.0.1" || host == "localhost" || host == "::1"
			},
		},
	}
}

func (b *WebTerminalBridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != b.token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch r.URL.Path {
	case "/", "":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(RenderWebTerminalHTML(b.token)))
	case "/ws":
		b.handleWebSocket(w, r)
	case "/assets/xterm.js":
		serveAsset(w, "application/javascript", webAsset("xterm.js"))
	case "/assets/xterm.css":
		serveAsset(w, "text/css", webAsset("xterm.css"))
	case "/assets/xterm-addon-fit.js":
		serveAsset(w, "application/javascript", webAsset("xterm-addon-fit.js"))
	default:
		http.NotFound(w, r)
	}
}

func (b *WebTerminalBridge) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if b.hasClient() {
		http.Error(w, "terminal is already open", http.StatusConflict)
		return
	}
	conn, err := b.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if !b.addClient(conn) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("terminal is already open"))
		_ = conn.Close()
		return
	}
	defer b.removeClient(conn)
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var message struct {
			Type    string `json:"type"`
			Data    string `json:"data"`
			Columns int    `json:"columns"`
			Rows    int    `json:"rows"`
		}
		if err := json.Unmarshal(data, &message); err != nil {
			continue
		}
		switch message.Type {
		case "input":
			if message.Data != "" {
				_ = b.relay.Send(SideClient, []byte(message.Data))
			}
		case ControlTypeResize:
			payload, _ := json.Marshal(map[string]any{"type": ControlTypeResize, "columns": message.Columns, "rows": message.Rows})
			_ = b.relay.Send(SideClient, append([]byte(ControlPrefix), payload...))
		}
	}
}

func (b *WebTerminalBridge) hasClient() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients) > 0
}

func (b *WebTerminalBridge) addClient(conn *websocket.Conn) bool {
	b.mu.Lock()
	if len(b.clients) > 0 {
		b.mu.Unlock()
		return false
	}
	b.clients[conn] = &sync.Mutex{}
	pending := append([][]byte(nil), b.pending...)
	b.pending = nil
	b.pendingBytes = 0
	b.mu.Unlock()
	for _, data := range pending {
		b.writeClient(conn, data)
	}
	return true
}

func (b *WebTerminalBridge) removeClient(conn *websocket.Conn) {
	b.mu.Lock()
	delete(b.clients, conn)
	b.mu.Unlock()
	_ = conn.Close()
}

func (b *WebTerminalBridge) relayLoop() {
	for {
		data, err := b.relay.Read(SideClient)
		if err != nil {
			closeOnce(b.stop)
			return
		}
		b.broadcast(data)
	}
}

func (b *WebTerminalBridge) broadcast(data []byte) {
	b.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(b.clients))
	for conn := range b.clients {
		clients = append(clients, conn)
	}
	if len(clients) == 0 {
		copied := append([]byte(nil), data...)
		b.pending = append(b.pending, copied)
		b.pendingBytes += len(copied)
		for len(b.pending) > 512 || b.pendingBytes > maxWebPendingBytes {
			b.pendingBytes -= len(b.pending[0])
			b.pending = b.pending[1:]
		}
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()
	for _, conn := range clients {
		b.writeClient(conn, data)
	}
}

func (b *WebTerminalBridge) writeClient(conn *websocket.Conn, data []byte) {
	b.mu.Lock()
	lock := b.clients[conn]
	b.mu.Unlock()
	if lock == nil {
		return
	}
	lock.Lock()
	err := conn.WriteMessage(websocket.BinaryMessage, data)
	lock.Unlock()
	if err != nil {
		b.removeClient(conn)
	}
}

func RunWebRelayClient(relayURL, sessionID, secret string) int {
	relay := NewRelayClient(relayURL, sessionID, secret)
	fmt.Fprintln(os.Stderr, "connecting to remote CLI...")
	if err := relay.Connect(SideClient); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	token, err := NewRandomURLToken(18)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	bridge := NewWebTerminalBridge(relay, token)
	server := &http.Server{Handler: bridge}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	if peer, err := relay.peer(SideClient); err == nil {
		keepaliveStop := make(chan struct{})
		defer close(keepaliveStop)
		go peer.KeepAlive(keepaliveStop)
	}
	go bridge.relayLoop()
	go func() { _ = server.Serve(listener) }()
	url := fmt.Sprintf("http://%s/?token=%s", listener.Addr().String(), token)
	fmt.Fprintln(os.Stderr, "connected successfully. Opening Controli terminal.")
	fmt.Fprintln(os.Stderr, "local terminal:", url)
	fmt.Fprintln(os.Stderr, "Keep this window open while the browser terminal is connected.")
	openBrowser(url)
	select {
	case <-bridge.stop:
	case <-stop:
	}
	relay.Close(SideClient)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	return 0
}

func closeOnce(ch chan struct{}) {
	defer func() { _ = recover() }()
	close(ch)
}

func UseWebTerminalByDefault() bool {
	return runtime.GOOS == "windows" || runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}
