package controli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	SideHost          = "host"
	SideClient        = "client"
	FinalCloseReason  = "controli-final-close"
	RelayPollInterval = 15 * time.Second
	MaxReconnectDelay = 5 * time.Second
	ControlPrefix     = "\x00CONTROLI:"
)

type RelayClient struct {
	RelayURL  string
	SessionID string
	Secret    string
	ClientID  string
	Timeout   time.Duration

	mu    sync.Mutex
	peers map[string]*RelayPeer
}

func NewRelayClient(relayURL, sessionID, secret string) *RelayClient {
	clientID, err := NewRandomURLToken(9)
	if err != nil {
		clientID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return &RelayClient{
		RelayURL:  strings.TrimRight(relayURL, "/"),
		SessionID: sessionID,
		Secret:    secret,
		ClientID:  clientID,
		Timeout:   35 * time.Second,
		peers:     map[string]*RelayPeer{},
	}
}

func (c *RelayClient) Connect(side string) error {
	_, err := c.peer(side)
	return err
}

func (c *RelayClient) Send(side string, data []byte) error {
	peer, err := c.peer(side)
	if err != nil {
		return err
	}
	return peer.Send(websocket.BinaryMessage, data)
}

func (c *RelayClient) Read(side string) ([]byte, error) {
	peer, err := c.peer(side)
	if err != nil {
		return nil, err
	}
	return peer.Read()
}

func (c *RelayClient) Close(side string) {
	c.mu.Lock()
	peer := c.peers[side]
	delete(c.peers, side)
	c.mu.Unlock()
	if peer != nil {
		peer.Close()
	}
	payload := map[string]any{"side": side}
	if side == SideClient {
		payload["client_id"] = c.ClientID
	}
	_ = c.post("/v1/close", payload, nil)
}

func (c *RelayClient) RegisterInvite(invite RelayToken) error {
	return c.post("/v1/invite/register", invite, nil)
}

func (c *RelayClient) ClaimInvite(code string) (RelayToken, error) {
	var token RelayToken
	err := c.post("/v1/invite/claim", map[string]string{"code": code}, &token)
	return token, err
}

func (c *RelayClient) Health() (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.httpURL()+"/health", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "controli-go")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("relay returned HTTP %d: %s", resp.StatusCode, string(data))
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *RelayClient) peer(side string) (*RelayPeer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if peer := c.peers[side]; peer != nil {
		return peer, nil
	}
	peer := &RelayPeer{client: c, side: side, closeCh: make(chan struct{})}
	if err := peer.connect(); err != nil {
		return nil, err
	}
	c.peers[side] = peer
	return peer, nil
}

func (c *RelayClient) post(path string, payload any, out any) error {
	var merged map[string]any
	if payloadBytes, err := json.Marshal(payload); err == nil {
		_ = json.Unmarshal(payloadBytes, &merged)
	}
	if merged == nil {
		merged = map[string]any{}
	}
	merged["session_id"] = c.SessionID
	merged["secret"] = c.Secret
	body, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.httpURL()+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "controli-go")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("relay returned HTTP %d: %s", resp.StatusCode, string(data))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *RelayClient) httpURL() string {
	if strings.HasPrefix(c.RelayURL, "wss://") {
		return "https://" + strings.TrimPrefix(c.RelayURL, "wss://")
	}
	if strings.HasPrefix(c.RelayURL, "ws://") {
		return "http://" + strings.TrimPrefix(c.RelayURL, "ws://")
	}
	return c.RelayURL
}

func (c *RelayClient) websocketURL(side string) (string, error) {
	parsed, err := url.Parse(c.RelayURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return "", errors.New("relay URL must start with ws:// or wss://")
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/v1/ws"
	}
	query := parsed.Query()
	query.Set("session_id", c.SessionID)
	query.Set("secret", c.Secret)
	query.Set("side", side)
	if side == SideClient {
		query.Set("client_id", c.ClientID)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

type RelayPeer struct {
	client *RelayClient
	side   string

	mu   sync.Mutex
	conn *websocket.Conn

	closeCh chan struct{}
}

func (p *RelayPeer) connect() error {
	return p.connectWithLimit(8)
}

func (p *RelayPeer) reconnect() error {
	return p.connectWithLimit(0)
}

func (p *RelayPeer) connectWithLimit(limit int) error {
	target, err := p.client.websocketURL(p.side)
	if err != nil {
		return err
	}
	var lastErr error
	delay := 250 * time.Millisecond
	for attempt := 0; limit == 0 || attempt < limit; attempt++ {
		select {
		case <-p.closeCh:
			return io.EOF
		default:
		}
		dialer := websocket.Dialer{HandshakeTimeout: p.client.Timeout}
		conn, _, err := dialer.Dial(target, http.Header{"User-Agent": []string{"controli-go"}})
		if err == nil {
			p.conn = conn
			return nil
		}
		lastErr = err
		select {
		case <-p.closeCh:
			return io.EOF
		case <-time.After(delay):
		}
		delay *= 2
		if delay > MaxReconnectDelay {
			delay = MaxReconnectDelay
		}
	}
	return fmt.Errorf("could not connect to relay after retries: %w", lastErr)
}

func (p *RelayPeer) Send(messageType int, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		if err := p.connect(); err != nil {
			return err
		}
	}
	if err := p.conn.WriteMessage(messageType, data); err != nil {
		_ = p.conn.Close()
		if err := p.reconnect(); err != nil {
			return err
		}
		return p.conn.WriteMessage(messageType, data)
	}
	return nil
}

func (p *RelayPeer) Read() ([]byte, error) {
	for {
		p.mu.Lock()
		conn := p.conn
		p.mu.Unlock()
		if conn == nil {
			p.mu.Lock()
			err := p.connect()
			p.mu.Unlock()
			if err != nil {
				return nil, err
			}
			continue
		}
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil, io.EOF
			}
			p.mu.Lock()
			_ = conn.Close()
			p.conn = nil
			err = p.reconnect()
			p.mu.Unlock()
			if err != nil {
				return nil, err
			}
			continue
		}
		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			return data, nil
		}
	}
}

func (p *RelayPeer) KeepAlive(stop <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			_ = p.Send(websocket.PingMessage, []byte("controli"))
		}
	}
}

func (p *RelayPeer) Close() {
	closeOnce(p.closeCh)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		return
	}
	_ = p.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, FinalCloseReason),
		time.Now().Add(time.Second),
	)
	_ = p.conn.Close()
	p.conn = nil
}
