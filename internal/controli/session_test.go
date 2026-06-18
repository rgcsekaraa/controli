package controli

import "testing"

func TestSameGuestReconnectKeepsApproval(t *testing.T) {
	gate := NewHostGate(HostModeFull, true)
	gate.GuestConnected(nil, "guest-a")
	gate.approved = true
	gate.approvedGuestID = "guest-a"

	gate.GuestDisconnected(nil, "guest-a", false)
	gate.GuestConnected(nil, "guest-a")

	if !gate.approved {
		t.Fatal("same guest reconnect should keep approval")
	}
	if gate.approvedGuestID != "guest-a" {
		t.Fatal("same guest reconnect should keep approved guest id")
	}
}

func TestDifferentGuestConnectionResetsApproval(t *testing.T) {
	gate := NewHostGate(HostModeFull, true)
	gate.GuestConnected(nil, "guest-a")
	gate.approved = true
	gate.approvedGuestID = "guest-a"

	gate.GuestDisconnected(nil, "guest-a", false)
	gate.GuestConnected(nil, "guest-b")

	if gate.approved {
		t.Fatal("different guest should require fresh approval")
	}
}

func TestFinalGuestDisconnectClearsApproval(t *testing.T) {
	gate := NewHostGate(HostModeFull, true)
	gate.GuestConnected(nil, "guest-a")
	gate.approved = true
	gate.approvedGuestID = "guest-a"

	gate.GuestDisconnected(nil, "guest-a", true)

	if gate.approved {
		t.Fatal("final guest disconnect should clear approval")
	}
	if gate.approvedGuestID != "" {
		t.Fatal("final guest disconnect should clear approved guest id")
	}
}

func TestApprovalDisabledKeepsGateOpen(t *testing.T) {
	gate := NewHostGate(HostModeFull, false)
	gate.approved = false

	gate.GuestConnected(nil, "guest-a")

	if !gate.approved {
		t.Fatal("approval should remain open when host disabled approval prompts")
	}
}
