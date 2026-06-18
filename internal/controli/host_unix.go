//go:build !windows

package controli

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/creack/pty"
)

func shellArgs(shell string) []string {
	name := strings.ToLower(filepath.Base(shell))
	switch name {
	case "powershell", "powershell.exe", "pwsh", "pwsh.exe":
		return []string{shell, "-NoLogo"}
	case "cmd", "cmd.exe":
		return []string{shell}
	default:
		return []string{shell, "-l"}
	}
}

func shellEnv() []string {
	env := os.Environ()
	if os.Getenv("TERM") == "" || os.Getenv("TERM") == "dumb" {
		env = append(env, "TERM=xterm-256color")
	}
	if os.Getenv("COLORTERM") == "" {
		env = append(env, "COLORTERM=truecolor")
	}
	if os.Getenv("CLICOLOR") == "" {
		env = append(env, "CLICOLOR=1")
	}
	return env
}

func RunHostRelayShell(relayURL, sessionID, secret, cwd, shell string) int {
	return RunHostRelayShellWithOptions(HostOptions{
		RelayURL:  relayURL,
		SessionID: sessionID,
		Secret:    secret,
		Cwd:       cwd,
		Shell:     shell,
		Mode:      HostModeFull,
	})
}

func RunHostRelayShellWithOptions(options HostOptions) int {
	shell := options.Shell
	if shell == "" {
		shell = DefaultShell()
	}
	command := exec.Command(shellArgs(shell)[0], shellArgs(shell)[1:]...)
	command.Dir = options.Cwd
	command.Env = shellEnv()
	tty, err := pty.StartWithSize(command, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to start PTY: " + err.Error() + "\n")
		return 1
	}
	defer func() { _ = tty.Close() }()

	relay := NewRelayClient(options.RelayURL, options.SessionID, options.Secret)
	if err := relay.Connect(SideHost); err != nil {
		_, _ = os.Stderr.WriteString("failed to connect relay: " + err.Error() + "\n")
		return 1
	}
	audit, err := OpenAuditLog(options.AuditLogPath)
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to open audit log: " + err.Error() + "\n")
		return 1
	}
	defer audit.Close()
	audit.Log("host_start", map[string]any{
		"session_id": options.SessionID,
		"workspace":  options.WorkspaceName,
		"cwd":        options.Cwd,
		"shell":      shell,
		"mode":       options.Mode,
	})
	stats := NewSessionStats()
	gate := NewHostGate(options.Mode, options.RequireApprove)
	stop := make(chan struct{})
	if peer, err := relay.peer(SideHost); err == nil {
		go peer.KeepAlive(stop)
	}
	if options.StatusInterval > 0 {
		go hostStatusLoop(stop, options.StatusInterval, stats)
	}

	go func() {
		buffer := make([]byte, ptyChunkSize)
		for {
			n, err := tty.Read(buffer)
			if n > 0 {
				stats.AddOutput(n)
				audit.Log("output", map[string]any{"bytes": n})
				_ = relay.Send(SideHost, buffer[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	go func() {
		for {
			data, err := relay.Read(SideHost)
			if err != nil {
				_ = command.Process.Signal(syscall.SIGTERM)
				return
			}
			if handleHostControl(tty, data, audit, gate) {
				continue
			}
			allowed, notice := gate.AllowInput(data, audit, options.AuditInput)
			if notice != "" {
				_ = relay.Send(SideHost, []byte(notice))
			}
			if !allowed {
				continue
			}
			stats.AddInput(len(data))
			_, _ = tty.Write(data)
		}
	}()

	err = command.Wait()
	close(stop)
	relay.Close(SideHost)
	audit.Log("host_stop", map[string]any{"stats": stats.Summary()})
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func handleHostControl(tty *os.File, data []byte, audit *AuditLog, gate *HostGate) bool {
	if !strings.HasPrefix(string(data), ControlPrefix) {
		return false
	}
	var payload ControlMessage
	if err := json.Unmarshal(data[len(ControlPrefix):], &payload); err != nil {
		return false
	}
	switch payload.Type {
	case ControlTypeResize:
		if payload.Columns <= 0 || payload.Rows <= 0 {
			return true
		}
		_ = pty.Setsize(tty, &pty.Winsize{Rows: payload.Rows, Cols: payload.Columns})
		audit.Log("resize", map[string]any{"columns": payload.Columns, "rows": payload.Rows})
	case ControlTypeGuestConnected:
		gate.GuestConnected(audit)
	case ControlTypeGuestDisconnected:
		gate.GuestDisconnected(audit)
	}
	return true
}
