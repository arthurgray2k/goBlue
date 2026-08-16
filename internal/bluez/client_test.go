package bluez

import (
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

func TestDevicePathFromAddress(t *testing.T) {
	adapterPath := dbus.ObjectPath("/org/bluez/hci0")
	address := "28:04:c6:73:e8:d6"
	expected := dbus.ObjectPath("/org/bluez/hci0/dev_28_04_C6_73_E8_D6")

	got := DevicePathFromAddress(adapterPath, address)
	if got != expected {
		t.Errorf("DevicePathFromAddress() = %s, want %s", got, expected)
	}
}

func TestAgentCallbacks(t *testing.T) {
	agent := NewAgent(nil, "NoInputNoOutput")
	if agent.capability != "NoInputNoOutput" {
		t.Errorf("expected capability NoInputNoOutput, got %s", agent.capability)
	}

	// Test default callbacks don't panic
	if err := agent.Release(); err != nil {
		t.Errorf("Release error: %v", err)
	}
	if err := agent.Cancel(); err != nil {
		t.Errorf("Cancel error: %v", err)
	}
	if err := agent.RequestAuthorization("/org/bluez/hci0/dev_11_22_33_44_55_66"); err != nil {
		t.Errorf("RequestAuthorization error: %v", err)
	}
	if err := agent.AuthorizeService("/org/bluez/hci0/dev_11_22_33_44_55_66", UUIDA2DPSink); err != nil {
		t.Errorf("AuthorizeService error: %v", err)
	}
	if err := agent.DisplayPasskey("/org/bluez/hci0/dev_11_22_33_44_55_66", 123456, 6); err != nil {
		t.Errorf("DisplayPasskey error: %v", err)
	}
}

func TestSetAdapterAlias(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("D-Bus not available, skipping live test")
	}
	defer client.Close()

	adapter, err := client.GetDefaultAdapter("")
	if err != nil {
		t.Skip("No adapter found, skipping live test")
	}

	origAlias := adapter.Alias
	defer func() {
		_ = client.SetAdapterAlias(adapter.Path, origAlias)
	}()

	// Register agent first
	agent := NewAgent(client.conn, "KeyboardDisplay")
	_ = agent.Register()
	defer agent.Unregister()

	// Test setting alias
	err = client.SetAdapterAlias(adapter.Path, "TestAlias")
	if err != nil {
		t.Fatalf("SetAdapterAlias failed: %v", err)
	}

	// Wait for kernel HCI management command to apply
	time.Sleep(150 * time.Millisecond)

	// Verify
	adapters, err := client.GetAdapters()
	if err != nil {
		t.Fatalf("GetAdapters failed: %v", err)
	}
	var found *Adapter
	for _, a := range adapters {
		if a.Path == adapter.Path {
			found = a
			break
		}
	}
	if found == nil || found.Alias != "TestAlias" {
		t.Errorf("expected alias TestAlias, got %v", found)
	}
}
