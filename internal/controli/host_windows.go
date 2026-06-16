//go:build windows

package controli

import (
	"fmt"
	"os"
)

func RunHostRelayShell(relayURL, sessionID, secret, cwd, shell string) int {
	fmt.Fprintln(os.Stderr, "Windows hosting is not enabled yet; Windows can join sessions now, and ConPTY hosting is the next Go milestone.")
	return 1
}
