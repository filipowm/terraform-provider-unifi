package device

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
)

func TestMergePortOverrides_preservesUnmanagedFields(t *testing.T) {
	// Current state from the controller: port 1 has QoS, storm control, etc.
	current := []unifi.DevicePortOverrides{
		{
			PortIDX:                      1,
			Name:                         "Port 1",
			PortProfileID:                "profile-abc",
			OpMode:                       "switch",
			PoeMode:                      "auto",
			AggregateNumPorts:            0,
			EgressRateLimitKbps:          1000,
			EgressRateLimitKbpsEnabled:   true,
			PriorityQueue1Level:          25,
			PriorityQueue2Level:          25,
			PriorityQueue3Level:          25,
			PriorityQueue4Level:          25,
			StormctrlBroadcastastEnabled: true,
			StormctrlBroadcastastRate:    500,
			Dot1XCtrl:                    "auto",
			Forward:                      "all",
			NATiveNetworkID:              "net-123",
		},
		{
			PortIDX:                    2,
			Name:                       "Port 2",
			OpMode:                     "aggregate",
			AggregateNumPorts:          4,
			EgressRateLimitKbps:        2000,
			EgressRateLimitKbpsEnabled: true,
		},
	}

	// Terraform manages port 1 only, changing its name and profile.
	managed := []unifi.DevicePortOverrides{
		{
			PortIDX:       1,
			Name:          "Updated Port 1",
			PortProfileID: "profile-xyz",
			OpMode:        "switch",
			PoeMode:       "off",
		},
	}

	result := mergePortOverrides(current, managed)

	if len(result) != 2 {
		t.Fatalf("expected 2 port overrides, got %d", len(result))
	}

	// Find port 1 in the result.
	var port1 *unifi.DevicePortOverrides
	var port2 *unifi.DevicePortOverrides
	for i := range result {
		switch result[i].PortIDX {
		case 1:
			port1 = &result[i]
		case 2:
			port2 = &result[i]
		}
	}

	if port1 == nil {
		t.Fatal("port 1 not found in result")
	}
	if port2 == nil {
		t.Fatal("port 2 not found in result")
	}

	// Verify Terraform-managed fields were updated for port 1.
	if port1.Name != "Updated Port 1" {
		t.Errorf("expected Name 'Updated Port 1', got %q", port1.Name)
	}
	if port1.PortProfileID != "profile-xyz" {
		t.Errorf("expected PortProfileID 'profile-xyz', got %q", port1.PortProfileID)
	}
	if port1.PoeMode != "off" {
		t.Errorf("expected PoeMode 'off', got %q", port1.PoeMode)
	}

	// Verify controller-managed fields were PRESERVED for port 1.
	if port1.EgressRateLimitKbps != 1000 {
		t.Errorf("expected EgressRateLimitKbps 1000, got %d", port1.EgressRateLimitKbps)
	}
	if !port1.EgressRateLimitKbpsEnabled {
		t.Error("expected EgressRateLimitKbpsEnabled to be true")
	}
	if port1.PriorityQueue1Level != 25 {
		t.Errorf("expected PriorityQueue1Level 25, got %d", port1.PriorityQueue1Level)
	}
	if port1.PriorityQueue2Level != 25 {
		t.Errorf("expected PriorityQueue2Level 25, got %d", port1.PriorityQueue2Level)
	}
	if !port1.StormctrlBroadcastastEnabled {
		t.Error("expected StormctrlBroadcastastEnabled to be true")
	}
	if port1.StormctrlBroadcastastRate != 500 {
		t.Errorf("expected StormctrlBroadcastastRate 500, got %d", port1.StormctrlBroadcastastRate)
	}
	if port1.Dot1XCtrl != "auto" {
		t.Errorf("expected Dot1XCtrl 'auto', got %q", port1.Dot1XCtrl)
	}
	if port1.Forward != "all" {
		t.Errorf("expected Forward 'all', got %q", port1.Forward)
	}
	if port1.NATiveNetworkID != "net-123" {
		t.Errorf("expected NATiveNetworkID 'net-123', got %q", port1.NATiveNetworkID)
	}

	// Verify port 2 (not managed by Terraform) was left completely unchanged.
	if port2.Name != "Port 2" {
		t.Errorf("expected port 2 Name 'Port 2', got %q", port2.Name)
	}
	if port2.OpMode != "aggregate" {
		t.Errorf("expected port 2 OpMode 'aggregate', got %q", port2.OpMode)
	}
	if port2.AggregateNumPorts != 4 {
		t.Errorf("expected port 2 AggregateNumPorts 4, got %d", port2.AggregateNumPorts)
	}
	if port2.EgressRateLimitKbps != 2000 {
		t.Errorf("expected port 2 EgressRateLimitKbps 2000, got %d", port2.EgressRateLimitKbps)
	}
}

func TestMergePortOverrides_newPortAdded(t *testing.T) {
	current := []unifi.DevicePortOverrides{
		{PortIDX: 1, Name: "Port 1", OpMode: "switch"},
	}

	// Terraform manages port 1 and adds a new port 3.
	managed := []unifi.DevicePortOverrides{
		{PortIDX: 1, Name: "Port 1 Updated", OpMode: "switch"},
		{PortIDX: 3, Name: "New Port 3", OpMode: "mirror"},
	}

	result := mergePortOverrides(current, managed)

	if len(result) != 2 {
		t.Fatalf("expected 2 port overrides, got %d", len(result))
	}

	// Verify the new port was added.
	found := false
	for _, r := range result {
		if r.PortIDX == 3 {
			found = true
			if r.Name != "New Port 3" {
				t.Errorf("expected new port Name 'New Port 3', got %q", r.Name)
			}
			if r.OpMode != "mirror" {
				t.Errorf("expected new port OpMode 'mirror', got %q", r.OpMode)
			}
		}
	}
	if !found {
		t.Error("new port 3 not found in result")
	}
}

func TestMergePortOverrides_emptyManaged(t *testing.T) {
	current := []unifi.DevicePortOverrides{
		{PortIDX: 1, Name: "Port 1", EgressRateLimitKbps: 1000},
		{PortIDX: 2, Name: "Port 2", EgressRateLimitKbps: 2000},
	}

	// No Terraform-managed overrides — all ports should be unchanged.
	result := mergePortOverrides(current, nil)

	if len(result) != 2 {
		t.Fatalf("expected 2 port overrides, got %d", len(result))
	}
	if result[0].EgressRateLimitKbps != 1000 {
		t.Errorf("expected port 1 EgressRateLimitKbps 1000, got %d", result[0].EgressRateLimitKbps)
	}
	if result[1].EgressRateLimitKbps != 2000 {
		t.Errorf("expected port 2 EgressRateLimitKbps 2000, got %d", result[1].EgressRateLimitKbps)
	}
}
