package controli

import (
	"strings"
	"testing"
)

func TestRelayTokenRoundTrip(t *testing.T) {
	token := RelayToken{
		Kind:      RelayTokenKind,
		Version:   1,
		SessionID: "session-1",
		Name:      "guest",
		RelayURL:  DefaultRelayURL,
		Secret:    "secret",
		ExpiresAt: "2026-06-17T01:00:00Z",
	}
	encoded, err := EncodeRelayToken(token)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeRelayToken(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.SessionID != token.SessionID || decoded.Secret != token.Secret || decoded.RelayURL != token.RelayURL {
		t.Fatalf("decoded token mismatch: %#v", decoded)
	}
}

func TestShortCodeNormalization(t *testing.T) {
	if got := NormalizeShortCode(" code: 123-45 67 "); got != "1234567" {
		t.Fatalf("NormalizeShortCode() = %q", got)
	}
	code, err := NewShortCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != ShortCodeLength {
		t.Fatalf("short code length = %d", len(code))
	}
}

func TestNoExpiryValueDoesNotExpire(t *testing.T) {
	if IsExpired(NoExpiryValue) {
		t.Fatal("never expiry should not expire")
	}
}

func TestWebTerminalHTMLUsesEmbeddedAssets(t *testing.T) {
	if len(webAsset("xterm.js")) == 0 {
		t.Fatal("xterm.js asset is empty")
	}
	html := RenderWebTerminalHTML("token-123")
	for _, want := range []string{"token-123", "/assets/xterm.js", "/assets/xterm.css", "/ws?token="} {
		if !strings.Contains(html, want) {
			t.Fatalf("RenderWebTerminalHTML missing %q", want)
		}
	}
}
