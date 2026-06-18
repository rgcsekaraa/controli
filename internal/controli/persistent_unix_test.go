//go:build !windows

package controli

import "testing"

func TestPersistentSessionNameUsesWorkspace(t *testing.T) {
	got := persistentSessionName(HostOptions{WorkspaceName: "Main Work"})
	if got != "controli-main-work" {
		t.Fatalf("persistentSessionName() = %q", got)
	}
}

func TestPersistentSessionNameUsesExplicitName(t *testing.T) {
	got := persistentSessionName(HostOptions{WorkspaceName: "main", PersistName: "Support_1"})
	if got != "support_1" {
		t.Fatalf("persistentSessionName() = %q", got)
	}
}

func TestShellCommandStringQuotesArgs(t *testing.T) {
	got := shellCommandString([]string{"/bin/zsh", "-l", "it's"})
	want := "'/bin/zsh' '-l' 'it'\\''s'"
	if got != want {
		t.Fatalf("shellCommandString() = %q, want %q", got, want)
	}
}
