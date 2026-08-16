package bluez

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// Device represents a Bluetooth remote device known to BlueZ.
type Device struct {
	Path        dbus.ObjectPath   `json:"path"`
	AdapterPath dbus.ObjectPath   `json:"adapter_path"`
	Address     string            `json:"address"`
	Name        string            `json:"name"`
	Alias       string            `json:"alias"`
	Class       uint32            `json:"class,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Paired      bool              `json:"paired"`
	Trusted     bool              `json:"trusted"`
	Connected   bool              `json:"connected"`
	Blocked     bool              `json:"blocked"`
	RSSI        int16             `json:"rssi,omitempty"`
	UUIDs       []string          `json:"uuids,omitempty"`
	AudioCaps   AudioCapabilities `json:"audio_capabilities"`
}

// NewDeviceFromProps constructs a Device from D-Bus object path and properties map.
func NewDeviceFromProps(objPath dbus.ObjectPath, props map[string]dbus.Variant) *Device {
	d := &Device{
		Path: objPath,
	}

	if v, ok := props["Adapter"]; ok {
		if p, ok := v.Value().(dbus.ObjectPath); ok {
			d.AdapterPath = p
		}
	}
	if v, ok := props["Address"]; ok {
		if s, ok := v.Value().(string); ok {
			d.Address = s
		}
	}
	if v, ok := props["Name"]; ok {
		if s, ok := v.Value().(string); ok {
			d.Name = s
		}
	}
	if v, ok := props["Alias"]; ok {
		if s, ok := v.Value().(string); ok {
			d.Alias = s
		}
	}
	if v, ok := props["Class"]; ok {
		if u, ok := v.Value().(uint32); ok {
			d.Class = u
		}
	}
	if v, ok := props["Icon"]; ok {
		if s, ok := v.Value().(string); ok {
			d.Icon = s
		}
	}
	if v, ok := props["Paired"]; ok {
		if b, ok := v.Value().(bool); ok {
			d.Paired = b
		}
	}
	if v, ok := props["Trusted"]; ok {
		if b, ok := v.Value().(bool); ok {
			d.Trusted = b
		}
	}
	if v, ok := props["Connected"]; ok {
		if b, ok := v.Value().(bool); ok {
			d.Connected = b
		}
	}
	if v, ok := props["Blocked"]; ok {
		if b, ok := v.Value().(bool); ok {
			d.Blocked = b
		}
	}
	if v, ok := props["RSSI"]; ok {
		if n, ok := v.Value().(int16); ok {
			d.RSSI = n
		}
	}
	if v, ok := props["UUIDs"]; ok {
		if uuids, ok := v.Value().([]string); ok {
			d.UUIDs = uuids
		}
	}

	d.AudioCaps = DetectAudioCapabilities(d.UUIDs)
	return d
}

// DisplayName returns Alias, Name, or Address as fallback.
func (d *Device) DisplayName() string {
	if d.Alias != "" {
		return d.Alias
	}
	if d.Name != "" {
		return d.Name
	}
	return d.Address
}

// StateString returns a concise primary state string.
func (d *Device) StateString() string {
	if d.Connected {
		return "connected"
	}
	if d.Paired {
		return "paired"
	}
	return "available"
}

// RSSIString returns formatted RSSI or "-" if not available.
func (d *Device) RSSIString() string {
	if d.RSSI == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", d.RSSI)
}

// TypeString returns a human-readable device category.
func (d *Device) TypeString() string {
	if d.Icon != "" {
		return d.Icon
	}
	// Major device class from Class of Device (CoD)
	major := (d.Class >> 8) & 0x1f
	switch major {
	case 1:
		return "computer"
	case 2:
		return "phone"
	case 3:
		return "network"
	case 4:
		return "audio/video"
	case 5:
		return "peripheral"
	case 6:
		return "imaging"
	case 7:
		return "wearable"
	case 8:
		return "toy"
	case 9:
		return "health"
	default:
		return "device"
	}
}
