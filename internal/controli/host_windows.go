//go:build windows

package controli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

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
	command := exec.Command(shell)
	command.Dir = options.Cwd
	command.Env = os.Environ()
	stdin, err := command.StdinPipe()
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to open stdin: " + err.Error() + "\n")
		return 1
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to open stdout: " + err.Error() + "\n")
		return 1
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to open stderr: " + err.Error() + "\n")
		return 1
	}
	if err := command.Start(); err != nil {
		_, _ = os.Stderr.WriteString("failed to start shell: " + err.Error() + "\n")
		return 1
	}
	relay := NewRelayClient(options.RelayURL, options.SessionID, options.Secret)
	if err := relay.Connect(SideHost); err != nil {
		_, _ = os.Stderr.WriteString("failed to connect relay: " + err.Error() + "\n")
		_ = command.Process.Kill()
		return 1
	}
	audit, err := OpenAuditLog(options.AuditLogPath)
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to open audit log: " + err.Error() + "\n")
		_ = command.Process.Kill()
		return 1
	}
	defer audit.Close()
	audit.Log("host_start", map[string]any{
		"session_id": options.SessionID,
		"workspace":  options.WorkspaceName,
		"cwd":        options.Cwd,
		"shell":      shell,
		"mode":       options.Mode,
		"backend":    "windows-stdio",
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
	go streamHostOutput(stdout, relay, audit, stats)
	go streamHostOutput(stderr, relay, audit, stats)
	go func() {
		for {
			data, err := relay.Read(SideHost)
			if err != nil {
				_ = command.Process.Kill()
				return
			}
			if handleWindowsHostControl(data, relay, audit, gate, options) {
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
			_, _ = stdin.Write(data)
		}
	}()
	err = command.Wait()
	close(stop)
	relay.Close(SideHost)
	audit.Log("host_stop", map[string]any{"stats": stats.Summary(), "backend": "windows-stdio"})
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func streamHostOutput(reader io.Reader, relay *RelayClient, audit *AuditLog, stats *SessionStats) {
	buffer := make([]byte, ptyChunkSize)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			stats.AddOutput(n)
			audit.Log("output", map[string]any{"bytes": n})
			_ = relay.Send(SideHost, buffer[:n])
		}
		if err != nil {
			return
		}
	}
}

func handleWindowsHostControl(data []byte, relay *RelayClient, audit *AuditLog, gate *HostGate, options HostOptions) bool {
	if !strings.HasPrefix(string(data), ControlPrefix) {
		return false
	}
	var payload ControlMessage
	if err := json.Unmarshal(data[len(ControlPrefix):], &payload); err != nil {
		return false
	}
	switch payload.Type {
	case ControlTypeGuestConnected:
		gate.GuestConnected(audit, payload.ClientID)
	case ControlTypeGuestDisconnected:
		gate.GuestDisconnected(audit, payload.ClientID, payload.Final)
	case ControlTypeResize:
		audit.Log("control_ignored", map[string]any{"backend": "windows-stdio", "type": payload.Type})
	case ControlTypeDownloadRequest:
		go handleDownloadRequest(relay, audit, options, payload)
	}
	return true
}
