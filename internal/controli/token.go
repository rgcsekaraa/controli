package controli

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

const (
	RelayTokenKind  = "controli-relay-token"
	SessionPrefix   = "ct1_"
	ShortCodeLength = 7
)

type RelayToken struct {
	Kind            string `json:"kind"`
	Version         int    `json:"version"`
	Code            string `json:"code,omitempty"`
	SessionID       string `json:"session_id"`
	Name            string `json:"name"`
	Room            string `json:"room,omitempty"`
	Mode            string `json:"mode,omitempty"`
	Transport       string `json:"transport,omitempty"`
	TunnelURL       string `json:"tunnel_url,omitempty"`
	RelayURL        string `json:"relay_url"`
	Secret          string `json:"secret"`
	ExpiresAt       string `json:"expires_at"`
	InviteExpiresAt string `json:"invite_expires_at,omitempty"`
}

func NewRandomURLToken(bytesLen int) (string, error) {
	raw := make([]byte, bytesLen)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func NewShortCode() (string, error) {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	value := int(raw[0])<<24 | int(raw[1])<<16 | int(raw[2])<<8 | int(raw[3])
	if value < 0 {
		value = -value
	}
	code := value % 10000000
	return leftPadInt(code, ShortCodeLength), nil
}

func leftPadInt(value, width int) string {
	out := ""
	for value > 0 {
		out = string(rune('0'+value%10)) + out
		value /= 10
	}
	for len(out) < width {
		out = "0" + out
	}
	return out
}

func EncodeRelayToken(token RelayToken) (string, error) {
	payload, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := writer.Write(payload); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return SessionPrefix + base64.RawURLEncoding.EncodeToString(compressed.Bytes()), nil
}

func DecodeRelayToken(value string) (RelayToken, error) {
	cleaned := strings.TrimSpace(strings.Trim(strings.TrimSpace(value), "'\""))
	if !strings.HasPrefix(cleaned, SessionPrefix) {
		for _, line := range strings.Split(cleaned, "\n") {
			stripped := strings.TrimSpace(strings.Trim(strings.TrimSpace(line), "'\""))
			if index := strings.Index(stripped, SessionPrefix); index >= 0 {
				cleaned = stripped[index:]
				break
			}
		}
	}
	if !strings.HasPrefix(cleaned, SessionPrefix) {
		return RelayToken{}, errors.New("that does not look like a Controli client code")
	}
	raw := regexp.MustCompile(`\s+`).ReplaceAllString(cleaned[len(SessionPrefix):], "")
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return RelayToken{}, errors.New("the client code looks incomplete or corrupted")
	}
	payload := decoded
	reader, err := zlib.NewReader(bytes.NewReader(decoded))
	if err == nil {
		payload, err = io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return RelayToken{}, err
		}
	}
	var token RelayToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return RelayToken{}, errors.New("the client code could not be read")
	}
	if token.Kind != RelayTokenKind {
		return RelayToken{}, errors.New("invalid Controli relay client code")
	}
	return token, nil
}

func NormalizeShortCode(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func IsExpired(value string) bool {
	t, err := time.Parse(time.RFC3339, value)
	return err != nil || time.Now().After(t)
}
