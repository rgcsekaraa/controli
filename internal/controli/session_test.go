package controli

import "testing"

func TestGuestConnectedResetsApprovalWhenApprovalRequired(t *testing.T) {
	gate := NewHostGate(HostModeFull, true)
	gate.approved = true
	gate.askedViewNotice = true

	gate.GuestConnected(nil)

	if gate.approved {
		t.Fatal("guest reconnect should reset approval")
	}
	if gate.askedViewNotice {
		t.Fatal("guest reconnect should reset view notice state")
	}
}

func TestGuestConnectedKeepsApprovalWhenApprovalDisabled(t *testing.T) {
	gate := NewHostGate(HostModeFull, false)
	gate.approved = false

	gate.GuestConnected(nil)

	if !gate.approved {
		t.Fatal("approval should remain open when host disabled approval prompts")
	}
}

func TestGuestConnectedRequiresApprovalModeEveryTime(t *testing.T) {
	gate := NewHostGate(HostModeApprove, false)
	gate.approved = true

	gate.GuestConnected(nil)

	if gate.approved {
		t.Fatal("approve mode should require approval after every reconnect")
	}
}
