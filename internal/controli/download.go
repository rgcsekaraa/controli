package controli

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDownloadDirName = "controli-drive"
	DefaultDownloadMax     = 100 * 1024 * 1024
	downloadChunkSize      = 48 * 1024
)

type controlSender interface {
	Send(side string, data []byte) error
}

func DefaultDownloadDir(workspacePath string) string {
	return filepath.Join(workspacePath, DefaultDownloadDirName)
}

func HashDownloadCode(value string) string {
	return HashJoinPassword(value)
}

func downloadCodeMatches(expectedHash, provided string) bool {
	expectedHash = strings.TrimSpace(expectedHash)
	provided = strings.TrimSpace(provided)
	if expectedHash == "" || provided == "" {
		return false
	}
	providedHash := HashDownloadCode(provided)
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(providedHash)) == 1
}

func cleanDownloadRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("download root is not configured")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func resolveDownloadPath(root, requested string) (string, error) {
	root, err := cleanDownloadRoot(root)
	if err != nil {
		return "", err
	}
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return "", fmt.Errorf("download path is required")
	}
	if filepath.IsAbs(requested) {
		return "", fmt.Errorf("download path must be relative to %s", DefaultDownloadDirName)
	}
	cleaned := filepath.Clean(requested)
	if cleaned == "." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return "", fmt.Errorf("download path cannot leave %s", DefaultDownloadDirName)
	}
	target := filepath.Join(root, cleaned)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("download path cannot leave %s", DefaultDownloadDirName)
	}
	return target, nil
}

func sendControl(sender controlSender, payload ControlMessage) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return sender.Send(SideHost, append([]byte(ControlPrefix), data...))
}

func handleDownloadRequest(sender controlSender, audit *AuditLog, options HostOptions, request ControlMessage) {
	id := strings.TrimSpace(request.ID)
	if id == "" {
		id = "download"
	}
	fail := func(message string) {
		audit.Log("download_denied", map[string]any{"id": id, "path": request.Path, "reason": message})
		_ = sendControl(sender, ControlMessage{Type: ControlTypeDownloadError, ID: id, Error: message})
	}
	if !options.Downloads {
		fail("downloads are disabled")
		return
	}
	root, err := cleanDownloadRoot(options.DownloadDir)
	if err != nil {
		fail(err.Error())
		return
	}
	target, err := resolveDownloadPath(root, request.Path)
	if err != nil {
		fail(err.Error())
		return
	}
	info, err := os.Stat(target)
	if err != nil {
		fail("file is not available")
		return
	}
	if info.IsDir() {
		fail("directories cannot be downloaded")
		return
	}
	max := options.DownloadMax
	if max <= 0 {
		max = DefaultDownloadMax
	}
	if info.Size() > max {
		fail(fmt.Sprintf("file exceeds download limit of %s", formatBytes(uint64(max))))
		return
	}
	rel, _ := filepath.Rel(root, target)
	authorizedBy := "host_prompt"
	if downloadCodeMatches(options.DownloadCodeHash, request.DownloadCode) {
		authorizedBy = "download_code"
		audit.Log("download_approved", map[string]any{"id": id, "path": target, "method": authorizedBy})
	} else {
		if options.DownloadCodeHash != "" && strings.TrimSpace(request.DownloadCode) != "" {
			audit.Log("download_code_denied", map[string]any{"id": id, "path": target})
		}
		if !promptHost(fmt.Sprintf("Allow guest to download %q from %s?", rel, DefaultDownloadDirName)) {
			fail("host denied download")
			return
		}
		audit.Log("download_approved", map[string]any{"id": id, "path": target, "method": authorizedBy})
	}
	file, err := os.Open(target)
	if err != nil {
		fail("file could not be opened")
		return
	}
	defer file.Close()
	audit.Log("download_start", map[string]any{"id": id, "path": target, "bytes": info.Size(), "method": authorizedBy})
	if err := sendControl(sender, ControlMessage{Type: ControlTypeDownloadStart, ID: id, Name: filepath.Base(target), Size: info.Size()}); err != nil {
		audit.Log("download_error", map[string]any{"id": id, "path": target, "error": err.Error()})
		return
	}
	buffer := make([]byte, downloadChunkSize)
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			encoded := base64.StdEncoding.EncodeToString(buffer[:n])
			if sendErr := sendControl(sender, ControlMessage{Type: ControlTypeDownloadChunk, ID: id, Data: encoded}); sendErr != nil {
				audit.Log("download_error", map[string]any{"id": id, "path": target, "error": sendErr.Error()})
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fail("file could not be read")
			return
		}
	}
	_ = sendControl(sender, ControlMessage{Type: ControlTypeDownloadDone, ID: id})
	audit.Log("download_done", map[string]any{"id": id, "path": target, "bytes": info.Size()})
}
