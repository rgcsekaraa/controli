package controli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type HostMode string

const (
	HostModeFull    HostMode = "full"
	HostModeView    HostMode = "view"
	HostModeApprove HostMode = "approve"
	ptyChunkSize             = 64 * 1024
)

const (
	ControlTypeResize            = "resize"
	ControlTypeGuestConnected    = "guest_connected"
	ControlTypeGuestDisconnected = "guest_disconnected"
)

type HostOptions struct {
	RelayURL       string
	SessionID      string
	Secret         string
	Cwd            string
	Shell          string
	WorkspaceName  string
	GuestName      string
	Mode           HostMode
	RequireApprove bool
	AuditLogPath   string
	AuditInput     bool
	StatusInterval time.Duration
	Persist        bool
	PersistName    string
}

type ControlMessage struct {
	Type     string `json:"type"`
	Columns  uint16 `json:"columns,omitempty"`
	Rows     uint16 `json:"rows,omitempty"`
	Text     string `json:"text,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Final    bool   `json:"final,omitempty"`
}

type SessionStats struct {
	startedAt time.Time
	inBytes   atomic.Uint64
	outBytes  atomic.Uint64
	lastUnix  atomic.Int64
}

func NewSessionStats() *SessionStats {
	stats := &SessionStats{startedAt: time.Now()}
	stats.Touch()
	return stats
}

func (s *SessionStats) AddInput(n int) {
	s.inBytes.Add(uint64(n))
	s.Touch()
}

func (s *SessionStats) AddOutput(n int) {
	s.outBytes.Add(uint64(n))
	s.Touch()
}

func (s *SessionStats) Touch() {
	s.lastUnix.Store(time.Now().Unix())
}

func (s *SessionStats) Summary() string {
	last := time.Unix(s.lastUnix.Load(), 0)
	return fmt.Sprintf(
		"duration=%s input=%s output=%s last_activity=%s",
		time.Since(s.startedAt).Truncate(time.Second),
		formatBytes(s.inBytes.Load()),
		formatBytes(s.outBytes.Load()),
		time.Since(last).Truncate(time.Second),
	)
}

type HostGate struct {
	mode            HostMode
	requireApprove  bool
	approved        bool
	askedViewNotice bool
	activeGuestID   string
	approvedGuestID string
	mu              sync.Mutex
}

func NewHostGate(mode HostMode, requireApprove bool) *HostGate {
	if mode == "" {
		mode = HostModeFull
	}
	return &HostGate{
		mode:           mode,
		requireApprove: requireApprove,
		approved:       !requireApprove && mode != HostModeApprove,
	}
}

func ParseHostMode(value string) (HostMode, error) {
	switch HostMode(strings.ToLower(strings.TrimSpace(value))) {
	case "", HostModeFull:
		return HostModeFull, nil
	case HostModeView:
		return HostModeView, nil
	case HostModeApprove:
		return HostModeApprove, nil
	default:
		return "", fmt.Errorf("mode must be one of: full, view, approve")
	}
}

func DefaultAuditLogPath(sessionID string) string {
	return filepath.Join(ControliHome(), "audit", sessionID+".jsonl")
}

func formatBytes(value uint64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%dB", value)
	}
	div, exp := uint64(unit), 0
	for n := value / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(value)/float64(div), "KMGTPE"[exp])
}

func (g *HostGate) AllowInput(data []byte, audit *AuditLog, auditInput bool) (bool, string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mode == HostModeView {
		audit.Log("input_blocked", map[string]any{"mode": g.mode, "bytes": len(data)})
		if g.askedViewNotice {
			return false, ""
		}
		g.askedViewNotice = true
		return false, "\r\n[controli] host is in view-only mode; input is disabled\r\n"
	}
	if !g.approved {
		if !promptHost("Allow guest to control this shell?") {
			audit.Log("control_denied", map[string]any{"bytes": len(data)})
			return false, "\r\n[controli] host denied control\r\n"
		}
		g.approved = true
		g.approvedGuestID = g.activeGuestID
		audit.Log("control_approved", map[string]any{"mode": g.mode})
	}
	fields := map[string]any{"bytes": len(data)}
	if auditInput {
		fields["text"] = string(data)
	}
	if g.mode == HostModeApprove {
		if !promptHost(fmt.Sprintf("Allow input %q?", string(data))) {
			audit.Log("input_denied", fields)
			return false, "\r\n[controli] host denied input\r\n"
		}
		audit.Log("input_approved", fields)
		return true, ""
	}
	audit.Log("input", fields)
	return true, ""
}

func (g *HostGate) GuestConnected(audit *AuditLog, clientID string) {
	g.mu.Lock()
	clientID = normalizeGuestID(clientID)
	sameApprovedGuest := clientID != "" && clientID == g.approvedGuestID && g.approved
	g.activeGuestID = clientID
	if !sameApprovedGuest {
		g.approved = !g.requireApprove && g.mode != HostModeApprove
		g.askedViewNotice = false
	}
	g.mu.Unlock()
	audit.Log("guest_connected", map[string]any{"mode": g.mode, "client_id": clientID, "approval_reused": sameApprovedGuest})
}

func (g *HostGate) GuestDisconnected(audit *AuditLog, clientID string, final bool) {
	g.mu.Lock()
	clientID = normalizeGuestID(clientID)
	if clientID == g.activeGuestID {
		g.activeGuestID = ""
	}
	if final && (clientID == "" || clientID == g.approvedGuestID) {
		g.approvedGuestID = ""
		g.approved = !g.requireApprove && g.mode != HostModeApprove
		g.askedViewNotice = false
	}
	g.mu.Unlock()
	audit.Log("guest_disconnected", map[string]any{"mode": g.mode, "client_id": clientID, "final": final})
}

func normalizeGuestID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "legacy-client"
	}
	return clientID
}

func promptHost(question string) bool {
	fmt.Fprintf(os.Stderr, "\n[controli] %s [y/N]: ", question)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func hostStatusLoop(stop <-chan struct{}, interval time.Duration, stats *SessionStats) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			fmt.Fprintln(os.Stderr, "[controli] status:", stats.Summary())
		}
	}
}
