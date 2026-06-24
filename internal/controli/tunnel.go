package controli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

const (
	TransportRelay  = "relay"
	TransportTunnel = "tunnel"
)

type TunnelHostOptions struct {
	HostOptions
	ListenAddr string
	PublicURL  string
}

type tunnelTerminalServer struct {
	token       string
	audit       *AuditLog
	stats       *SessionStats
	gate        *HostGate
	auditInput  bool
	options     HostOptions
	stop        chan struct{}
	input       func([]byte) error
	resize      func(uint16, uint16)
	clients     map[*websocket.Conn]*sync.Mutex
	clientIDs   map[*websocket.Conn]string
	clientsByID map[string]*websocket.Conn
	mu          sync.Mutex
	upgrader    websocket.Upgrader
}

func newTunnelTerminalServer(token string, audit *AuditLog, stats *SessionStats, gate *HostGate, options HostOptions, input func([]byte) error, resize func(uint16, uint16)) *tunnelTerminalServer {
	return &tunnelTerminalServer{
		token:       token,
		audit:       audit,
		stats:       stats,
		gate:        gate,
		auditInput:  options.AuditInput,
		options:     options,
		stop:        make(chan struct{}),
		input:       input,
		resize:      resize,
		clients:     map[*websocket.Conn]*sync.Mutex{},
		clientIDs:   map[*websocket.Conn]string{},
		clientsByID: map[string]*websocket.Conn{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return r.URL.Query().Get("token") == token
			},
		},
	}
}

func (s *tunnelTerminalServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != s.token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch r.URL.Path {
	case "/", "":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(RenderWebTerminalHTML(s.token)))
	case "/ws":
		s.handleWebSocket(w, r)
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

func (s *tunnelTerminalServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	clientID := normalizeGuestID(r.URL.Query().Get("client_id"))
	if s.hasDifferentClient(clientID) {
		http.Error(w, "session already has an active guest", http.StatusConflict)
		return
	}
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if !s.addClient(conn, clientID) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("session already has an active guest"))
		_ = conn.Close()
		return
	}
	s.audit.Log("guest_connected", map[string]any{"remote_addr": r.RemoteAddr, "client_id": clientID})
	defer func() {
		s.audit.Log("guest_disconnected", map[string]any{"remote_addr": r.RemoteAddr, "client_id": clientID})
		s.removeClient(conn, false)
	}()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var message struct {
			Type         string `json:"type"`
			Data         string `json:"data"`
			Columns      int    `json:"columns"`
			Rows         int    `json:"rows"`
			ID           string `json:"id"`
			Path         string `json:"path"`
			DownloadCode string `json:"download_code"`
		}
		if err := json.Unmarshal(data, &message); err != nil {
			continue
		}
		switch message.Type {
		case "input":
			if message.Data == "" {
				continue
			}
			input := []byte(message.Data)
			allowed, notice := s.gate.AllowInput(input, s.audit, s.auditInput)
			if notice != "" {
				s.writeClient(conn, []byte(notice))
			}
			if !allowed {
				continue
			}
			s.stats.AddInput(len(input))
			_ = s.input(input)
		case ControlTypeResize:
			if message.Columns > 0 && message.Rows > 0 {
				s.resize(uint16(message.Columns), uint16(message.Rows))
				s.audit.Log("resize", map[string]any{"columns": message.Columns, "rows": message.Rows})
			}
		case ControlTypeDownloadRequest:
			go handleDownloadRequest(tunnelDownloadSender{server: s, conn: conn}, s.audit, s.hostOptions(), ControlMessage{
				Type:         message.Type,
				ID:           message.ID,
				Path:         message.Path,
				DownloadCode: message.DownloadCode,
			})
		}
	}
}

func (s *tunnelTerminalServer) hostOptions() HostOptions {
	return s.options
}

type tunnelDownloadSender struct {
	server *tunnelTerminalServer
	conn   *websocket.Conn
}

func (s tunnelDownloadSender) Send(side string, data []byte) error {
	s.server.writeClient(s.conn, data)
	return nil
}

func (s *tunnelTerminalServer) hasDifferentClient(clientID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.clients) == 0 {
		return false
	}
	existing := s.clientsByID[clientID]
	return existing == nil
}

func (s *tunnelTerminalServer) addClient(conn *websocket.Conn, clientID string) bool {
	s.mu.Lock()
	clientID = normalizeGuestID(clientID)
	if existing := s.clientsByID[clientID]; existing != nil {
		delete(s.clients, existing)
		delete(s.clientIDs, existing)
		delete(s.clientsByID, clientID)
		_ = existing.Close()
	} else if len(s.clients) > 0 {
		s.mu.Unlock()
		return false
	}
	s.clients[conn] = &sync.Mutex{}
	s.clientIDs[conn] = clientID
	s.clientsByID[clientID] = conn
	s.mu.Unlock()
	s.gate.GuestConnected(s.audit, clientID)
	return true
}

func (s *tunnelTerminalServer) removeClient(conn *websocket.Conn, final bool) {
	s.mu.Lock()
	clientID, existed := s.clientIDs[conn]
	if existed {
		delete(s.clients, conn)
		delete(s.clientIDs, conn)
		if s.clientsByID[clientID] == conn {
			delete(s.clientsByID, clientID)
		}
	}
	s.mu.Unlock()
	_ = conn.Close()
	if existed {
		s.gate.GuestDisconnected(s.audit, clientID, final)
	}
}

func (s *tunnelTerminalServer) broadcast(data []byte) {
	s.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(s.clients))
	for conn := range s.clients {
		clients = append(clients, conn)
	}
	s.mu.Unlock()
	for _, conn := range clients {
		s.writeClient(conn, data)
	}
}

func (s *tunnelTerminalServer) writeClient(conn *websocket.Conn, data []byte) {
	s.mu.Lock()
	lock := s.clients[conn]
	s.mu.Unlock()
	if lock == nil {
		return
	}
	lock.Lock()
	err := conn.WriteMessage(websocket.BinaryMessage, data)
	lock.Unlock()
	if err != nil {
		s.removeClient(conn, false)
	}
}

func serveTunnelTerminal(options TunnelHostOptions, server *tunnelTerminalServer) int {
	listenAddr := strings.TrimSpace(options.ListenAddr)
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8765"
	}
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer listener.Close()
	httpServer := &http.Server{Handler: server}
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopSignal
		closeOnce(server.stop)
	}()
	go func() {
		err := httpServer.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "server error:", err)
			closeOnce(server.stop)
		}
	}()
	fmt.Println("tunnel terminal is ready")
	fmt.Println("local_service:", "http://"+listener.Addr().String())
	if options.PublicURL != "" {
		fmt.Println("public_url:", JoinTunnelURL(options.PublicURL, options.Secret))
	}
	<-server.stop
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	return 0
}

func JoinTunnelURL(publicURL, secret string) string {
	parsed, err := url.Parse(strings.TrimSpace(publicURL))
	if err != nil {
		return publicURL
	}
	query := parsed.Query()
	query.Set("token", secret)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func RunTunnelJoin(publicURL, secret string) int {
	target := JoinTunnelURL(publicURL, secret)
	fmt.Fprintln(os.Stderr, "opening tunnel terminal:", target)
	openBrowser(target)
	return 0
}
