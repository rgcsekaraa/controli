//go:build !windows

package controli

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func newHostCommand(options HostOptions, shell string) (*exec.Cmd, string, string) {
	if options.Persist {
		sessionName := persistentSessionName(options)
		if tmuxPath, err := exec.LookPath("tmux"); err == nil {
			args := []string{"new-session", "-A", "-s", sessionName}
			if strings.TrimSpace(options.Cwd) != "" {
				args = append(args, "-c", options.Cwd)
			}
			args = append(args, shellCommandString(shellArgs(shell)))
			command := exec.Command(tmuxPath, args...)
			command.Dir = options.Cwd
			command.Env = shellEnv()
			return command, "tmux", sessionName
		}
	}
	args := shellArgs(shell)
	command := exec.Command(args[0], args[1:]...)
	command.Dir = options.Cwd
	command.Env = shellEnv()
	return command, "pty", ""
}

func persistentSessionName(options HostOptions) string {
	if value := sanitizeTmuxName(options.PersistName); value != "" {
		return value
	}
	if value := sanitizeTmuxName(options.WorkspaceName); value != "" {
		return "controli-" + value
	}
	if value := sanitizeTmuxName(filepath.Base(options.Cwd)); value != "" && value != "." && value != string(filepath.Separator) {
		return "controli-" + value
	}
	return "controli-session"
}

func sanitizeTmuxName(value string) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(unicode.ToLower(r))
		case r == '-' || r == '_' || r == '.':
			builder.WriteRune(r)
		case unicode.IsSpace(r):
			builder.WriteByte('-')
		}
	}
	return strings.Trim(builder.String(), "-_.")
}

func shellCommandString(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func persistentBackendMessage(backend, name string) string {
	if backend != "tmux" {
		return "pty"
	}
	return fmt.Sprintf("tmux:%s", name)
}
