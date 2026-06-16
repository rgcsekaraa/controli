package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgcsekaraa/controli/internal/controli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var code int
	switch os.Args[1] {
	case "relay":
		code = cmdRelay(os.Args[2:])
	case "host":
		code = cmdHost(os.Args[2:])
	case "join":
		code = cmdJoin(os.Args[2:])
	default:
		usage()
		code = 2
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  controli relay configure --url wss://<relay>")
	fmt.Fprintln(os.Stderr, "  controli relay status")
	fmt.Fprintln(os.Stderr, "  controli host share --workspace <name>")
	fmt.Fprintln(os.Stderr, "  controli join <code>")
}

func cmdRelay(args []string) int {
	if len(args) < 1 {
		usage()
		return 2
	}
	switch args[0] {
	case "configure":
		flags := flag.NewFlagSet("relay configure", flag.ContinueOnError)
		urlValue := flags.String("url", "", "relay URL")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if !strings.HasPrefix(*urlValue, "wss://") && !strings.HasPrefix(*urlValue, "ws://") {
			fmt.Fprintln(os.Stderr, "error: relay URL must start with wss:// or ws://")
			return 1
		}
		state, err := controli.LoadState()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		state.Relay.URL = strings.TrimRight(*urlValue, "/")
		if err := controli.SaveState(state); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("relay configured:", state.Relay.URL)
		return 0
	case "status":
		state, err := controli.LoadState()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		if state.Relay.URL == "" {
			fmt.Println("relay_url: (not configured)")
			return 0
		}
		fmt.Println("relay_url:", state.Relay.URL)
		return 0
	default:
		usage()
		return 2
	}
}

func cmdHost(args []string) int {
	if len(args) < 1 || args[0] != "share" {
		usage()
		return 2
	}
	flags := flag.NewFlagSet("host share", flag.ContinueOnError)
	workspaceName := flags.String("workspace", "", "workspace name")
	relayURL := flags.String("relay-url", "", "relay URL")
	name := flags.String("name", "guest", "guest name")
	minutes := flags.Int("minutes", 120, "invite lifetime in minutes")
	shell := flags.String("shell", "", "shell path")
	printOnly := flags.Bool("print-only", false, "print code without starting shell")
	longCode := flags.Bool("long-code", false, "print full self-contained code")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if *workspaceName == "" {
		fmt.Fprintln(os.Stderr, "error: --workspace is required")
		return 1
	}
	state, err := controli.LoadState()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	workspace, ok := state.Workspaces[*workspaceName]
	if !ok {
		fmt.Fprintln(os.Stderr, "error: unknown workspace:", *workspaceName)
		return 1
	}
	activeRelayURL := strings.TrimRight(firstNonEmpty(*relayURL, state.Relay.URL), "/")
	if activeRelayURL == "" {
		fmt.Fprintln(os.Stderr, "error: relay URL is not configured")
		return 1
	}
	sessionID, err := controli.NewRandomURLToken(12)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	secret, err := controli.NewRandomURLToken(32)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	expiresAt := time.Now().UTC().Add(time.Duration(*minutes) * time.Minute).Truncate(time.Second).Format(time.RFC3339)
	token := controli.RelayToken{
		Kind:      controli.RelayTokenKind,
		Version:   1,
		SessionID: sessionID,
		Name:      *name,
		RelayURL:  activeRelayURL,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}
	if *longCode {
		encoded, err := controli.EncodeRelayToken(token)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println(encoded)
	} else {
		code, err := registerShortInvite(activeRelayURL, token)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("Send this 7-digit code to the guest (treat it like a password):")
		fmt.Println(code)
		fmt.Println()
		fmt.Println("expires_at:", expiresAt)
		fmt.Println("If short-code lookup fails, print a full code with: controli host share --workspace <name> --long-code")
	}
	if *printOnly {
		return 0
	}
	path := expandHome(workspace.Path)
	activeShell := firstNonEmpty(*shell, workspace.Shell, controli.DefaultShell())
	fmt.Println("relay session is ready; send the code to the guest and keep this process running")
	return controli.RunHostRelayShell(activeRelayURL, sessionID, secret, path, activeShell)
}

func registerShortInvite(relayURL string, token controli.RelayToken) (string, error) {
	client := controli.NewRelayClient(relayURL, token.SessionID, token.Secret)
	var lastErr error
	for i := 0; i < 10; i++ {
		code, err := controli.NewShortCode()
		if err != nil {
			return "", err
		}
		token.Code = code
		token.InviteExpiresAt = time.Now().UTC().Add(15 * time.Minute).Truncate(time.Second).Format(time.RFC3339)
		if err := client.RegisterInvite(token); err != nil {
			lastErr = err
			continue
		}
		return code, nil
	}
	return "", fmt.Errorf("could not register short invite code: %w", lastErr)
}

func cmdJoin(args []string) int {
	flags := flag.NewFlagSet("join", flag.ContinueOnError)
	relayURL := flags.String("relay-url", "", "relay URL for short code")
	webTerminal := flags.Bool("web-terminal", false, "open the local browser terminal")
	console := flags.Bool("console", false, "force direct console rendering")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	code := ""
	if flags.NArg() > 0 {
		code = flags.Arg(0)
	} else {
		fmt.Print("paste client code: ")
		_, _ = fmt.Scanln(&code)
	}
	token, err := resolveJoinToken(code, firstNonEmpty(*relayURL, controli.DefaultRelayURL))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if controli.IsExpired(token.ExpiresAt) {
		fmt.Fprintln(os.Stderr, "error: client code is expired")
		return 1
	}
	if *webTerminal || (controli.UseWebTerminalByDefault() && !*console) {
		return controli.RunWebRelayClient(token.RelayURL, token.SessionID, token.Secret)
	}
	return controli.RunConsoleRelayClient(token.RelayURL, token.SessionID, token.Secret)
}

func resolveJoinToken(value, relayURL string) (controli.RelayToken, error) {
	if strings.Contains(value, controli.SessionPrefix) {
		return controli.DecodeRelayToken(value)
	}
	code := controli.NormalizeShortCode(value)
	if len(code) != controli.ShortCodeLength {
		return controli.RelayToken{}, fmt.Errorf("short code must be %d digits", controli.ShortCodeLength)
	}
	client := controli.NewRelayClient(relayURL, "invite", "invite")
	return client.ClaimInvite(code)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func expandHome(value string) string {
	if value == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(value, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, value[2:])
	}
	return value
}
