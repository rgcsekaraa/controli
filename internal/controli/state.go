package controli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const DefaultRelayURL = "wss://controli-relay.rgcsekaraa.workers.dev"

type State struct {
	Relay      RelayState           `json:"relay"`
	Workspaces map[string]Workspace `json:"workspaces"`
}

type RelayState struct {
	URL string `json:"url"`
}

type Workspace struct {
	Path  string `json:"path"`
	Shell string `json:"shell"`
}

func ControliHome() string {
	if value := os.Getenv("CONTROLI_HOME"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".controli"
	}
	return filepath.Join(home, ".controli")
}

func StatePath() string {
	return filepath.Join(ControliHome(), "state.json")
}

func LoadState() (State, error) {
	state := State{Workspaces: map[string]Workspace{}}
	data, err := os.ReadFile(StatePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	if state.Workspaces == nil {
		state.Workspaces = map[string]Workspace{}
	}
	return state, nil
}

func SaveState(state State) error {
	if state.Workspaces == nil {
		state.Workspaces = map[string]Workspace{}
	}
	path := StatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func DefaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" && runtime.GOOS != "windows" {
		return shell
	}
	if runtime.GOOS == "windows" {
		if shell := os.Getenv("ComSpec"); shell != "" {
			return shell
		}
		return "cmd.exe"
	}
	return "/bin/sh"
}
