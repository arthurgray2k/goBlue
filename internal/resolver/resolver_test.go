package resolver

import (
	"strings"
	"testing"
	"time"

	"github.com/arthurgray2k/goBlue/internal/bluez"
	"github.com/godbus/dbus/v5"
)

type mockBlueZClient struct {
	devices []*bluez.Device
}

func (m *mockBlueZClient) Close() error { return nil }
func (m *mockBlueZClient) GetAdapters() ([]*bluez.Adapter, error) {
	return []*bluez.Adapter{{Path: "/org/bluez/hci0", ID: "hci0"}}, nil
}
func (m *mockBlueZClient) GetDefaultAdapter(pref string) (*bluez.Adapter, error) {
	return &bluez.Adapter{Path: "/org/bluez/hci0", ID: "hci0"}, nil
}
func (m *mockBlueZClient) SetAdapterAlias(p dbus.ObjectPath, a string) error { return nil }
func (m *mockBlueZClient) GetDevices(adapterPath dbus.ObjectPath) ([]*bluez.Device, error) {
	return m.devices, nil
}
func (m *mockBlueZClient) GetDevice(devicePath dbus.ObjectPath) (*bluez.Device, error) {
	for _, d := range m.devices {
		if d.Path == devicePath {
			return d, nil
		}
	}
	return nil, &DeviceNotFoundError{Query: string(devicePath)}
}
func (m *mockBlueZClient) StartDiscovery(p dbus.ObjectPath) error { return nil }
func (m *mockBlueZClient) StopDiscovery(p dbus.ObjectPath) error  { return nil }
func (m *mockBlueZClient) Scan(p dbus.ObjectPath, t time.Duration, f func(*bluez.Device)) ([]*bluez.Device, error) {
	return m.devices, nil
}
func (m *mockBlueZClient) Pair(p dbus.ObjectPath) error                     { return nil }
func (m *mockBlueZClient) CancelPairing(p dbus.ObjectPath) error            { return nil }
func (m *mockBlueZClient) Connect(p dbus.ObjectPath) error                  { return nil }
func (m *mockBlueZClient) Disconnect(p dbus.ObjectPath) error               { return nil }
func (m *mockBlueZClient) SetTrusted(p dbus.ObjectPath, trusted bool) error { return nil }
func (m *mockBlueZClient) RemoveDevice(a, d dbus.ObjectPath) error          { return nil }

func TestResolveDevice(t *testing.T) {
	devices := []*bluez.Device{
		{
			Path:      "/org/bluez/hci0/dev_AA_BB_CC_DD_EE_FF",
			Address:   "AA:BB:CC:DD:EE:FF",
			Name:      "Sony WH-1000XM5",
			Alias:     "Living Room Headphones",
			Connected: true,
			Paired:    true,
		},
		{
			Path:      "/org/bluez/hci0/dev_11_22_33_44_55_66",
			Address:   "11:22:33:44:55:66",
			Name:      "Galaxy Buds",
			Alias:     "Galaxy Buds",
			Connected: false,
			Paired:    true,
		},
		{
			Path:      "/org/bluez/hci0/dev_22_33_44_55_66_77",
			Address:   "22:33:44:55:66:77",
			Name:      "Bluetooth Speaker",
			Alias:     "Bluetooth Speaker",
			Connected: false,
			Paired:    false,
		},
		{
			Path:      "/org/bluez/hci0/dev_33_44_55_66_77_88",
			Address:   "33:44:55:66:77:88",
			Name:      "Bluetooth Speaker", // Duplicate name
			Alias:     "Bluetooth Speaker",
			Connected: false,
			Paired:    false,
		},
	}

	client := &mockBlueZClient{devices: devices}
	adapterPath := dbus.ObjectPath("/org/bluez/hci0")

	t.Run("Resolve by 1-based index", func(t *testing.T) {
		// Index 1 should be the connected Sony headphones (since SortDevices puts connected first)
		dev, err := ResolveDevice(client, adapterPath, "1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Address != "AA:BB:CC:DD:EE:FF" {
			t.Errorf("expected index 1 to be Sony (AA:BB:CC:DD:EE:FF), got %s (%s)", dev.DisplayName(), dev.Address)
		}
	})

	t.Run("Resolve index out of range", func(t *testing.T) {
		_, err := ResolveDevice(client, adapterPath, "99")
		if err == nil {
			t.Fatalf("expected error for index out of range, got nil")
		}
	})

	t.Run("Resolve by exact MAC address uppercase", func(t *testing.T) {
		dev, err := ResolveDevice(client, adapterPath, "AA:BB:CC:DD:EE:FF")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Address != "AA:BB:CC:DD:EE:FF" {
			t.Errorf("expected AA:BB:CC:DD:EE:FF, got %s", dev.Address)
		}
	})

	t.Run("Resolve by exact MAC address lowercase with hyphens", func(t *testing.T) {
		dev, err := ResolveDevice(client, adapterPath, "11-22-33-44-55-66")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Address != "11:22:33:44:55:66" {
			t.Errorf("expected 11:22:33:44:55:66, got %s", dev.Address)
		}
	})

	t.Run("Resolve by unique substring (short name)", func(t *testing.T) {
		dev, err := ResolveDevice(client, adapterPath, "sony")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Address != "AA:BB:CC:DD:EE:FF" {
			t.Errorf("expected Sony, got %s", dev.DisplayName())
		}
	})

	t.Run("Resolve by MAC suffix", func(t *testing.T) {
		dev, err := ResolveDevice(client, adapterPath, "55:66")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Address != "11:22:33:44:55:66" {
			t.Errorf("expected Galaxy Buds, got %s", dev.DisplayName())
		}
	})

	t.Run("Resolve by alias", func(t *testing.T) {
		dev, err := ResolveDevice(client, adapterPath, "Living Room Headphones")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Address != "AA:BB:CC:DD:EE:FF" {
			t.Errorf("expected AA:BB:CC:DD:EE:FF, got %s", dev.Address)
		}
	})

	t.Run("Resolve ambiguous duplicate names", func(t *testing.T) {
		_, err := ResolveDevice(client, adapterPath, "Bluetooth Speaker")
		if err == nil {
			t.Fatalf("expected AmbiguousDeviceError, got nil")
		}
		ambigErr, ok := err.(*AmbiguousDeviceError)
		if !ok {
			t.Fatalf("expected *AmbiguousDeviceError, got %T", err)
		}
		if len(ambigErr.Matches) != 2 {
			t.Errorf("expected 2 matches, got %d", len(ambigErr.Matches))
		}
		if !strings.Contains(ambigErr.Error(), "multiple devices found") {
			t.Errorf("error message missing multiple devices notice: %s", ambigErr.Error())
		}
	})

	t.Run("Resolve non-existent device", func(t *testing.T) {
		_, err := ResolveDevice(client, adapterPath, "Unknown Gadget")
		if err == nil {
			t.Fatalf("expected DeviceNotFoundError, got nil")
		}
		if _, ok := err.(*DeviceNotFoundError); !ok {
			t.Errorf("expected *DeviceNotFoundError, got %T", err)
		}
	})

	t.Run("Resolve empty identifier", func(t *testing.T) {
		_, err := ResolveDevice(client, adapterPath, "   ")
		if err == nil {
			t.Fatalf("expected error for empty identifier")
		}
	})
}
