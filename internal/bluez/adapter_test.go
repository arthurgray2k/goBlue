package bluez

import (
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestNewAdapterFromProps(t *testing.T) {
	props := map[string]dbus.Variant{
		"Address":      dbus.MakeVariant("38:7A:0E:12:20:7C"),
		"Name":         dbus.MakeVariant("mint-external"),
		"Alias":        dbus.MakeVariant("mint-controller"),
		"Powered":      dbus.MakeVariant(true),
		"Discoverable": dbus.MakeVariant(false),
		"Pairable":     dbus.MakeVariant(true),
		"Discovering":  dbus.MakeVariant(false),
		"UUIDs": dbus.MakeVariant([]string{
			"0000110a-0000-1000-8000-00805f9b34fb",
		}),
	}

	adapter := NewAdapterFromProps("/org/bluez/hci0", props)

	if adapter.ID != "hci0" {
		t.Errorf("expected ID hci0, got %s", adapter.ID)
	}
	if adapter.Address != "38:7A:0E:12:20:7C" {
		t.Errorf("expected address 38:7A:0E:12:20:7C, got %s", adapter.Address)
	}
	if adapter.DisplayName() != "mint-controller" {
		t.Errorf("expected DisplayName 'mint-controller', got %s", adapter.DisplayName())
	}
	if !adapter.Powered {
		t.Errorf("expected Powered true")
	}
	if !adapter.Pairable {
		t.Errorf("expected Pairable true")
	}
	if adapter.Discovering {
		t.Errorf("expected Discovering false")
	}
}
