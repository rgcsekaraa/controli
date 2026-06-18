//go:build windows

package controli

import (
	"io"
	"os"
	"os/exec"
)

func RunHostTunnelShellWithOptions(options TunnelHostOptions) int {
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
		"transport":  TransportTunnel,
		"backend":    "windows-stdio",
	})
	stats := NewSessionStats()
	gate := NewHostGate(options.Mode, options.RequireApprove)
	server := newTunnelTerminalServer(
		options.Secret,
		audit,
		stats,
		gate,
		options.AuditInput,
		func(data []byte) error {
			_, err := stdin.Write(data)
			return err
		},
		func(uint16, uint16) {},
	)
	if options.StatusInterval > 0 {
		go hostStatusLoop(server.stop, options.StatusInterval, stats)
	}
	go streamTunnelOutput(stdout, server, audit, stats)
	go streamTunnelOutput(stderr, server, audit, stats)
	go func() {
		_ = command.Wait()
		audit.Log("host_stop", map[string]any{"stats": stats.Summary(), "transport": TransportTunnel, "backend": "windows-stdio"})
		closeOnce(server.stop)
	}()
	return serveTunnelTerminal(options, server)
}

func streamTunnelOutput(reader io.Reader, server *tunnelTerminalServer, audit *AuditLog, stats *SessionStats) {
	buffer := make([]byte, ptyChunkSize)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			stats.AddOutput(n)
			audit.Log("output", map[string]any{"bytes": n})
			server.broadcast(buffer[:n])
		}
		if err != nil {
			closeOnce(server.stop)
			return
		}
	}
}
