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

const (
	ptyChunkSize = 64 * 1024
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
	if shell == "" {
		shell = DefaultShell()
	}
	command := exec.Command(shellArgs(shell)[0], shellArgs(shell)[1:]...)
	command.Dir = cwd
	command.Env = shellEnv()
	tty, err := pty.StartWithSize(command, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to start PTY: " + err.Error() + "\n")
		return 1
	}
	defer func() { _ = tty.Close() }()

	relay := NewRelayClient(relayURL, sessionID, secret)
	if err := relay.Connect(SideHost); err != nil {
		_, _ = os.Stderr.WriteString("failed to connect relay: " + err.Error() + "\n")
		return 1
	}
	stop := make(chan struct{})
	if peer, err := relay.peer(SideHost); err == nil {
		go peer.KeepAlive(stop)
	}

	go func() {
		buffer := make([]byte, ptyChunkSize)
		for {
			n, err := tty.Read(buffer)
			if n > 0 {
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
			if handleHostControl(tty, data) {
				continue
			}
			_, _ = tty.Write(data)
		}
	}()

	err = command.Wait()
	close(stop)
	relay.Close(SideHost)
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func handleHostControl(tty *os.File, data []byte) bool {
	if !strings.HasPrefix(string(data), ControlPrefix) {
		return false
	}
	var payload struct {
		Type    string `json:"type"`
		Columns uint16 `json:"columns"`
		Rows    uint16 `json:"rows"`
	}
	if err := json.Unmarshal(data[len(ControlPrefix):], &payload); err != nil {
		return false
	}
	if payload.Type == "resize" && payload.Columns > 0 && payload.Rows > 0 {
		_ = pty.Setsize(tty, &pty.Winsize{Rows: payload.Rows, Cols: payload.Columns})
	}
	return true
}
