package connector

import (
	"fmt"
	"testing"
	"time"

	"github.com/arthurgray2k/goBlue/internal/bluez"
	"github.com/godbus/dbus/v5"
)

type mockClient struct {
	devices       []*bluez.Device
	connectedDevs map[dbus.ObjectPath]bool
	failConnect   bool
}

func newMockClient(devices []*bluez.Device) *mockClient {
	connMap := make(map[dbus.ObjectPath]bool)
	for _, d := range devices {
		if d.Connected {
			connMap[d.Path] = true
		}
	}
	return &mockClient{devices: devices, connectedDevs: connMap}
}

func (m *mockClient) Close() error { return nil }
func (m *mockClient) GetAdapters() ([]*bluez.Adapter, error) {
	return []*bluez.Adapter{{Path: "/org/bluez/hci0", ID: "hci0"}}, nil
}
func (m *mockClient) GetDefaultAdapter(p string) (*bluez.Adapter, error) {
	return &bluez.Adapter{Path: "/org/bluez/hci0", ID: "hci0"}, nil
}
func (m *mockClient) SetAdapterAlias(p dbus.ObjectPath, a string) error { return nil }
func (m *mockClient) GetDevices(adapterPath dbus.ObjectPath) ([]*bluez.Device, error) {
	return m.devices, nil
}
func (m *mockClient) GetDevice(devicePath dbus.ObjectPath) (*bluez.Device, error) {
	for _, d := range m.devices {
		if d.Path == devicePath {
			return d, nil
		}
	}
	return nil, fmt.Errorf("device not found")
}
func (m *mockClient) StartDiscovery(p dbus.ObjectPath) error { return nil }
func (m *mockClient) StopDiscovery(p dbus.ObjectPath) error  { return nil }
func (m *mockClient) Scan(p dbus.ObjectPath, t time.Duration, f func(*bluez.Device)) ([]*bluez.Device, error) {
	return m.devices, nil
}
func (m *mockClient) Pair(p dbus.ObjectPath) error          { return nil }
func (m *mockClient) CancelPairing(p dbus.ObjectPath) error { return nil }
func (m *mockClient) Connect(p dbus.ObjectPath) error {
	if m.failConnect {
		return fmt.Errorf("simulated connection failure")
	}
	m.connectedDevs[p] = true
	for _, d := range m.devices {
		if d.Path == p {
			d.Connected = true
		}
	}
	return nil
}
func (m *mockClient) Disconnect(p dbus.ObjectPath) error {
	delete(m.connectedDevs, p)
	for _, d := range m.devices {
		if d.Path == p {
			d.Connected = false
		}
	}
	return nil
}
func (m *mockClient) SetTrusted(p dbus.ObjectPath, trusted bool) error { return nil }
func (m *mockClient) RemoveDevice(a, d dbus.ObjectPath) error          { return nil }

func makeTestDevices() []*bluez.Device {
	return []*bluez.Device{
		{
			Path:      "/org/bluez/hci0/dev_AA_BB_CC_DD_EE_FF",
			Address:   "AA:BB:CC:DD:EE:FF",
			Name:      "Sony WH-1000XM5",
			Connected: false,
			AudioCaps: bluez.AudioCapabilities{HasA2DPSink: true},
		},
		{
			Path:      "/org/bluez/hci0/dev_11_22_33_44_55_66",
			Address:   "11:22:33:44:55:66",
			Name:      "Galaxy Buds",
			Connected: true,
			AudioCaps: bluez.AudioCapabilities{HasA2DPSink: true},
		},
	}
}

func TestConnectorConnect(t *testing.T) {
	adapterPath := dbus.ObjectPath("/org/bluez/hci0")

	t.Run("Connect disconnected device", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		res, err := connector.Connect(adapterPath, "Sony WH-1000XM5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AlreadyConnected {
			t.Errorf("expected AlreadyConnected false")
		}
		if !res.HasA2DPSink {
			t.Errorf("expected HasA2DPSink true")
		}
	})

	t.Run("Connect already connected device", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		res, err := connector.Connect(adapterPath, "Galaxy Buds")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.AlreadyConnected {
			t.Errorf("expected AlreadyConnected true")
		}
	})

	t.Run("Connect failure error handling", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		client.failConnect = true
		connector := NewConnector(client)

		_, err := connector.Connect(adapterPath, "AA:BB:CC:DD:EE:FF")
		if err == nil {
			t.Fatalf("expected error on failed connection, got nil")
		}
	})

	t.Run("Connect non-existent device", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		_, err := connector.Connect(adapterPath, "NonExistent")
		if err == nil {
			t.Fatalf("expected error for non-existent device")
		}
	})
}

func TestConnectorDisconnect(t *testing.T) {
	adapterPath := dbus.ObjectPath("/org/bluez/hci0")

	t.Run("Disconnect active device", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		res, err := connector.Disconnect(adapterPath, "Galaxy Buds")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AlreadyDisconnected {
			t.Errorf("expected AlreadyDisconnected false")
		}
	})

	t.Run("Disconnect already disconnected device", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		res, err := connector.Disconnect(adapterPath, "Sony WH-1000XM5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.AlreadyDisconnected {
			t.Errorf("expected AlreadyDisconnected true")
		}
	})
}

func TestEnsureConnected(t *testing.T) {
	adapterPath := dbus.ObjectPath("/org/bluez/hci0")

	t.Run("EnsureConnected already connected", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		dev, err := connector.EnsureConnected(adapterPath, "Galaxy Buds")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dev.Connected {
			t.Errorf("expected device to be connected")
		}
	})

	t.Run("EnsureConnected currently disconnected", func(t *testing.T) {
		client := newMockClient(makeTestDevices())
		connector := NewConnector(client)

		dev, err := connector.EnsureConnected(adapterPath, "Sony WH-1000XM5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dev.Connected {
			t.Errorf("expected device to become connected")
		}
	})
}
