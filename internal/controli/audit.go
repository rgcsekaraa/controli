package controli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditLog struct {
	mu   sync.Mutex
	file *os.File
}

func OpenAuditLog(path string) (*AuditLog, error) {
	if path == "" {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &AuditLog{file: file}, nil
}

func (a *AuditLog) Close() {
	if a == nil || a.file == nil {
		return
	}
	_ = a.file.Close()
}

func (a *AuditLog) Log(event string, fields map[string]any) {
	if a == nil || a.file == nil {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	fields["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	fields["event"] = event
	data, err := json.Marshal(fields)
	if err != nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, _ = a.file.Write(append(data, '\n'))
}
