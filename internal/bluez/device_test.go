package bluez

import (
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestNewDeviceFromProps(t *testing.T) {
	props := map[string]dbus.Variant{
		"Address":   dbus.MakeVariant("AA:BB:CC:DD:EE:FF"),
		"Name":      dbus.MakeVariant("Sony WH-1000XM5"),
		"Alias":     dbus.MakeVariant("Sony Headphones"),
		"Paired":    dbus.MakeVariant(true),
		"Trusted":   dbus.MakeVariant(true),
		"Connected": dbus.MakeVariant(false),
		"RSSI":      dbus.MakeVariant(int16(-55)),
		"Class":     dbus.MakeVariant(uint32(2360324)),
		"Icon":      dbus.MakeVariant("audio-headset"),
		"UUIDs": dbus.MakeVariant([]string{
			"0000110b-0000-1000-8000-00805f9b34fb",
		}),
	}

	dev := NewDeviceFromProps("/org/bluez/hci0/dev_AA_BB_CC_DD_EE_FF", props)

	if dev.Address != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected address AA:BB:CC:DD:EE:FF, got %s", dev.Address)
	}
	if dev.DisplayName() != "Sony Headphones" {
		t.Errorf("expected DisplayName 'Sony Headphones', got %s", dev.DisplayName())
	}
	if dev.StateString() != "paired" {
		t.Errorf("expected StateString 'paired', got %s", dev.StateString())
	}
	if dev.RSSIString() != "-55" {
		t.Errorf("expected RSSI '-55', got %s", dev.RSSIString())
	}
	if dev.TypeString() != "audio-headset" {
		t.Errorf("expected TypeString 'audio-headset', got %s", dev.TypeString())
	}
	if !dev.AudioCaps.HasA2DPSink {
		t.Errorf("expected HasA2DPSink true")
	}

	// Test connected state
	props["Connected"] = dbus.MakeVariant(true)
	devConn := NewDeviceFromProps("/org/bluez/hci0/dev_AA_BB_CC_DD_EE_FF", props)
	if devConn.StateString() != "connected" {
		t.Errorf("expected StateString 'connected', got %s", devConn.StateString())
	}

	// Test available state (unpaired)
	props["Paired"] = dbus.MakeVariant(false)
	props["Connected"] = dbus.MakeVariant(false)
	devAvail := NewDeviceFromProps("/org/bluez/hci0/dev_AA_BB_CC_DD_EE_FF", props)
	if devAvail.StateString() != "available" {
		t.Errorf("expected StateString 'available', got %s", devAvail.StateString())
	}
}

func TestDeviceDisplayNameFallback(t *testing.T) {
	// Only address
	dev1 := NewDeviceFromProps("/org/bluez/hci0/dev_11_22_33_44_55_66", map[string]dbus.Variant{
		"Address": dbus.MakeVariant("11:22:33:44:55:66"),
	})
	if dev1.DisplayName() != "11:22:33:44:55:66" {
		t.Errorf("expected address as fallback, got %s", dev1.DisplayName())
	}

	// Name present, no Alias
	dev2 := NewDeviceFromProps("/org/bluez/hci0/dev_11_22_33_44_55_66", map[string]dbus.Variant{
		"Address": dbus.MakeVariant("11:22:33:44:55:66"),
		"Name":    dbus.MakeVariant("DeviceNameOnly"),
	})
	if dev2.DisplayName() != "DeviceNameOnly" {
		t.Errorf("expected Name fallback, got %s", dev2.DisplayName())
	}
}
