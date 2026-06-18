package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	case "update":
		code = cmdUpdate(os.Args[2:])
	default:
		usage()
		code = 2
	}
	os.Exit(code)
}

func cmdUpdate(args []string) int {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	repo := flags.String("repo", controli.DefaultGitHubRepo, "GitHub repo in owner/name form")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if err := controli.UpdateFromLatestRelease(*repo, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  controli relay configure --url wss://<relay>")
	fmt.Fprintln(os.Stderr, "  controli relay status")
	fmt.Fprintln(os.Stderr, "  controli relay deploy")
	fmt.Fprintln(os.Stderr, "  controli host share --workspace <name>")
	fmt.Fprintln(os.Stderr, "  controli host tunnel --workspace <name> --public-url https://<host>")
	fmt.Fprintln(os.Stderr, "  controli join <code>")
	fmt.Fprintln(os.Stderr, "  controli update")
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
		health, err := controli.NewRelayClient(state.Relay.URL, "health", "health").Health()
		if err != nil {
			fmt.Println("health: unreachable")
			fmt.Println("error:", err)
			return 1
		}
		fmt.Println("health:", health["ok"])
		if service, ok := health["service"].(string); ok {
			fmt.Println("service:", service)
		}
		return 0
	case "deploy":
		dir := filepath.Join("infra", "cloudflare-relay")
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
			fmt.Fprintln(os.Stderr, "error: relay deploy must be run from the Controli source checkout")
			return 1
		}
		command := exec.Command("npm", "run", "deploy")
		command.Dir = dir
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		return 0
	default:
		usage()
		return 2
	}
}

func cmdHost(args []string) int {
	if len(args) < 1 {
		usage()
		return 2
	}
	if args[0] == "tunnel" {
		return cmdHostTunnel(args[1:])
	}
	if args[0] != "share" {
		usage()
		return 2
	}
	flags := flag.NewFlagSet("host share", flag.ContinueOnError)
	workspaceName := flags.String("workspace", "", "workspace name")
	room := flags.String("room", "", "room name shown to the guest")
	relayURL := flags.String("relay-url", "", "relay URL")
	name := flags.String("name", "guest", "guest name")
	minutes := flags.Int("minutes", 120, "invite lifetime in minutes")
	shell := flags.String("shell", "", "shell path")
	printOnly := flags.Bool("print-only", false, "print code without starting shell")
	longCode := flags.Bool("long-code", false, "print full self-contained code")
	modeValue := flags.String("mode", "full", "permission mode: full, view, approve")
	approve := flags.Bool("approve", true, "ask host before guest control starts")
	auditLog := flags.String("audit-log", "", "audit log path, or off")
	auditInput := flags.Bool("audit-input", false, "record typed input in the audit log")
	statusInterval := flags.Duration("status-interval", 0, "print host session status on an interval, for example 30s")
	persist := flags.Bool("persist", true, "keep the shell in a persistent host session when supported")
	persistName := flags.String("persist-name", "", "stable persistent shell name")
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
	mode, err := controli.ParseHostMode(*modeValue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
	expiresAt := expiryValue(*minutes)
	token := controli.RelayToken{
		Kind:      controli.RelayTokenKind,
		Version:   1,
		SessionID: sessionID,
		Name:      *name,
		Room:      firstNonEmpty(*room, *workspaceName),
		Mode:      string(mode),
		Transport: controli.TransportRelay,
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
		code, err := registerShortInvite(activeRelayURL, token, *minutes)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("Send this 7-digit code to the guest (treat it like a password):")
		fmt.Println(code)
		fmt.Println()
		fmt.Println("expires_at:", printableExpiry(expiresAt))
		fmt.Println("If short-code lookup fails, print a full code with: controli host share --workspace <name> --long-code")
	}
	if *printOnly {
		return 0
	}
	path := expandHome(workspace.Path)
	activeShell := firstNonEmpty(*shell, workspace.Shell, controli.DefaultShell())
	activeAuditLog := strings.TrimSpace(*auditLog)
	if activeAuditLog == "" {
		activeAuditLog = controli.DefaultAuditLogPath(sessionID)
	}
	if strings.EqualFold(activeAuditLog, "off") || strings.EqualFold(activeAuditLog, "none") {
		activeAuditLog = ""
	}
	fmt.Println("relay session is ready; send the code to the guest and keep this process running")
	fmt.Println("room:", firstNonEmpty(*room, *workspaceName))
	fmt.Println("permission_mode:", mode)
	fmt.Println("persistent_shell:", *persist)
	if activeAuditLog != "" {
		fmt.Println("audit_log:", activeAuditLog)
	}
	return controli.RunHostRelayShellWithOptions(controli.HostOptions{
		RelayURL:       activeRelayURL,
		SessionID:      sessionID,
		Secret:         secret,
		Cwd:            path,
		Shell:          activeShell,
		WorkspaceName:  *workspaceName,
		GuestName:      *name,
		Mode:           mode,
		RequireApprove: *approve,
		AuditLogPath:   activeAuditLog,
		AuditInput:     *auditInput,
		StatusInterval: *statusInterval,
		Persist:        *persist,
		PersistName:    *persistName,
	})
}

func cmdHostTunnel(args []string) int {
	flags := flag.NewFlagSet("host tunnel", flag.ContinueOnError)
	workspaceName := flags.String("workspace", "", "workspace name")
	room := flags.String("room", "", "room name shown to the guest")
	relayURL := flags.String("relay-url", "", "short-code API relay URL")
	publicURL := flags.String("public-url", "", "public Cloudflare Tunnel URL")
	listenAddr := flags.String("listen", "127.0.0.1:8765", "local HTTP service address for cloudflared")
	name := flags.String("name", "guest", "guest name")
	minutes := flags.Int("minutes", 1440, "invite lifetime in minutes")
	shell := flags.String("shell", "", "shell path")
	printOnly := flags.Bool("print-only", false, "print code without starting shell")
	longCode := flags.Bool("long-code", false, "print full self-contained code")
	modeValue := flags.String("mode", "full", "permission mode: full, view, approve")
	approve := flags.Bool("approve", true, "ask host before guest control starts")
	auditLog := flags.String("audit-log", "", "audit log path, or off")
	auditInput := flags.Bool("audit-input", false, "record typed input in the audit log")
	statusInterval := flags.Duration("status-interval", 0, "print host session status on an interval, for example 30s")
	persist := flags.Bool("persist", true, "keep the shell in a persistent host session when supported")
	persistName := flags.String("persist-name", "", "stable persistent shell name")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *workspaceName == "" {
		fmt.Fprintln(os.Stderr, "error: --workspace is required")
		return 1
	}
	if strings.TrimSpace(*publicURL) == "" {
		fmt.Fprintln(os.Stderr, "error: --public-url is required for tunnel mode")
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
	activeRelayURL := strings.TrimRight(firstNonEmpty(*relayURL, state.Relay.URL, controli.DefaultRelayURL), "/")
	mode, err := controli.ParseHostMode(*modeValue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
	expiresAt := expiryValue(*minutes)
	token := controli.RelayToken{
		Kind:      controli.RelayTokenKind,
		Version:   1,
		SessionID: sessionID,
		Name:      *name,
		Room:      firstNonEmpty(*room, *workspaceName),
		Mode:      string(mode),
		Transport: controli.TransportTunnel,
		TunnelURL: strings.TrimRight(*publicURL, "/"),
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
		code, err := registerShortInvite(activeRelayURL, token, *minutes)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("Send this 7-digit code to the guest (treat it like a password):")
		fmt.Println(code)
		fmt.Println()
		fmt.Println("expires_at:", printableExpiry(expiresAt))
		fmt.Println("transport:", controli.TransportTunnel)
	}
	if *printOnly {
		return 0
	}
	path := expandHome(workspace.Path)
	activeShell := firstNonEmpty(*shell, workspace.Shell, controli.DefaultShell())
	activeAuditLog := strings.TrimSpace(*auditLog)
	if activeAuditLog == "" {
		activeAuditLog = controli.DefaultAuditLogPath(sessionID)
	}
	if strings.EqualFold(activeAuditLog, "off") || strings.EqualFold(activeAuditLog, "none") {
		activeAuditLog = ""
	}
	fmt.Println("tunnel session is ready; keep cloudflared and this process running")
	fmt.Println("room:", firstNonEmpty(*room, *workspaceName))
	fmt.Println("permission_mode:", mode)
	fmt.Println("persistent_shell:", *persist)
	fmt.Println("listen:", *listenAddr)
	if activeAuditLog != "" {
		fmt.Println("audit_log:", activeAuditLog)
	}
	return controli.RunHostTunnelShellWithOptions(controli.TunnelHostOptions{
		HostOptions: controli.HostOptions{
			RelayURL:       activeRelayURL,
			SessionID:      sessionID,
			Secret:         secret,
			Cwd:            path,
			Shell:          activeShell,
			WorkspaceName:  *workspaceName,
			GuestName:      *name,
			Mode:           mode,
			RequireApprove: *approve,
			AuditLogPath:   activeAuditLog,
			AuditInput:     *auditInput,
			StatusInterval: *statusInterval,
			Persist:        *persist,
			PersistName:    *persistName,
		},
		ListenAddr: *listenAddr,
		PublicURL:  strings.TrimRight(*publicURL, "/"),
	})
}

func registerShortInvite(relayURL string, token controli.RelayToken, minutes int) (string, error) {
	client := controli.NewRelayClient(relayURL, token.SessionID, token.Secret)
	var lastErr error
	for i := 0; i < 10; i++ {
		code, err := controli.NewShortCode()
		if err != nil {
			return "", err
		}
		token.Code = code
		token.InviteExpiresAt = shortInviteExpiryValue(minutes)
		if err := client.RegisterInvite(token); err != nil {
			lastErr = err
			continue
		}
		return code, nil
	}
	return "", fmt.Errorf("could not register short invite code: %w", lastErr)
}

func expiryValue(minutes int) string {
	if minutes <= 0 {
		return controli.NoExpiryValue
	}
	return time.Now().UTC().Add(time.Duration(minutes) * time.Minute).Truncate(time.Second).Format(time.RFC3339)
}

func shortInviteExpiryValue(minutes int) string {
	if minutes <= 0 {
		return controli.NoExpiryValue
	}
	if minutes > 15 {
		minutes = 15
	}
	return time.Now().UTC().Add(time.Duration(minutes) * time.Minute).Truncate(time.Second).Format(time.RFC3339)
}

func printableExpiry(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), controli.NoExpiryValue) {
		return "never"
	}
	return value
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
	if token.Room != "" {
		fmt.Fprintln(os.Stderr, "room:", token.Room)
	}
	if token.Mode != "" {
		fmt.Fprintln(os.Stderr, "permission_mode:", token.Mode)
	}
	if token.Transport == controli.TransportTunnel {
		if token.TunnelURL == "" {
			fmt.Fprintln(os.Stderr, "error: tunnel code is missing a public URL")
			return 1
		}
		return controli.RunTunnelJoin(token.TunnelURL, token.Secret)
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
