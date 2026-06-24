package controli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type recordingControlSender struct {
	messages []ControlMessage
}

func (s *recordingControlSender) Send(side string, data []byte) error {
	if !bytes.HasPrefix(data, []byte(ControlPrefix)) {
		return nil
	}
	var message ControlMessage
	if err := json.Unmarshal(data[len(ControlPrefix):], &message); err != nil {
		return err
	}
	s.messages = append(s.messages, message)
	return nil
}

func TestResolveDownloadPathStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	got, err := resolveDownloadPath(root, filepath.Join("reports", "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "reports", "out.txt")
	if got != want {
		t.Fatalf("resolveDownloadPath() = %q, want %q", got, want)
	}
}

func TestResolveDownloadPathRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"..", "../secret.txt", filepath.Join("reports", "..", "..", "secret.txt")} {
		if got, err := resolveDownloadPath(root, path); err == nil {
			t.Fatalf("resolveDownloadPath(%q) = %q, want error", path, got)
		}
	}
}

func TestResolveDownloadPathRejectsAbsolutePath(t *testing.T) {
	root := t.TempDir()
	absolute := filepath.Join(root, "file.txt")
	if got, err := resolveDownloadPath(root, absolute); err == nil {
		t.Fatalf("resolveDownloadPath(%q) = %q, want error", absolute, got)
	}
}

func TestDownloadRequestWithValidCodeSkipsHostPrompt(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	sender := &recordingControlSender{}
	handleDownloadRequest(sender, nil, HostOptions{
		Downloads:        true,
		DownloadDir:      root,
		DownloadCodeHash: HashDownloadCode("MK_RGC_2026"),
	}, ControlMessage{
		ID:           "download-1",
		Path:         "report.txt",
		DownloadCode: "MK_RGC_2026",
	})
	if !hasDownloadMessage(sender.messages, ControlTypeDownloadStart) || !hasDownloadMessage(sender.messages, ControlTypeDownloadDone) {
		t.Fatalf("valid download code should complete download, got %#v", sender.messages)
	}
	if hasDownloadMessage(sender.messages, ControlTypeDownloadError) {
		t.Fatalf("valid download code should not return an error, got %#v", sender.messages)
	}
}

func TestDownloadRequestWithWrongCodeCanFallBackToHostApproval(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "y\n", func() {
		sender := &recordingControlSender{}
		handleDownloadRequest(sender, nil, HostOptions{
			Downloads:        true,
			DownloadDir:      root,
			DownloadCodeHash: HashDownloadCode("MK_RGC_2026"),
		}, ControlMessage{
			ID:           "download-1",
			Path:         "report.txt",
			DownloadCode: "wrong",
		})
		if !hasDownloadMessage(sender.messages, ControlTypeDownloadDone) {
			t.Fatalf("host approval should complete download, got %#v", sender.messages)
		}
	})
}

func TestDownloadRequestWithWrongCodeCanBeDeniedByHost(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "n\n", func() {
		sender := &recordingControlSender{}
		handleDownloadRequest(sender, nil, HostOptions{
			Downloads:        true,
			DownloadDir:      root,
			DownloadCodeHash: HashDownloadCode("MK_RGC_2026"),
		}, ControlMessage{
			ID:           "download-1",
			Path:         "report.txt",
			DownloadCode: "wrong",
		})
		if !hasDownloadMessage(sender.messages, ControlTypeDownloadError) {
			t.Fatalf("host denial should return download error, got %#v", sender.messages)
		}
	})
}

func hasDownloadMessage(messages []ControlMessage, messageType string) bool {
	for _, message := range messages {
		if message.Type == messageType {
			return true
		}
	}
	return false
}

func withStdin(t *testing.T, input string, run func()) {
	t.Helper()
	original := os.Stdin
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := write.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = read
	defer func() {
		os.Stdin = original
		_ = read.Close()
	}()
	run()
}
