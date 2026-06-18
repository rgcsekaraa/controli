//go:build !windows

package controli

import (
	"errors"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func RunHostTunnelShellWithOptions(options TunnelHostOptions) int {
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
		"transport":  TransportTunnel,
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
			_, err := tty.Write(data)
			return err
		},
		func(columns, rows uint16) {
			_ = pty.Setsize(tty, &pty.Winsize{Rows: rows, Cols: columns})
		},
	)
	if options.StatusInterval > 0 {
		go hostStatusLoop(server.stop, options.StatusInterval, stats)
	}
	go func() {
		buffer := make([]byte, ptyChunkSize)
		for {
			n, err := tty.Read(buffer)
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
	}()
	go func() {
		err := command.Wait()
		audit.Log("host_stop", map[string]any{"stats": stats.Summary(), "transport": TransportTunnel})
		closeOnce(server.stop)
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				audit.Log("shell_exit", map[string]any{"code": exitErr.ExitCode()})
			}
		}
	}()
	return serveTunnelTerminal(options, server)
}
