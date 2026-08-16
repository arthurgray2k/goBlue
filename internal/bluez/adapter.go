package bluez

import (
	"path"

	"github.com/godbus/dbus/v5"
)

// Adapter represents a local Bluetooth controller/adapter managed by BlueZ (e.g. hci0).
type Adapter struct {
	Path         dbus.ObjectPath `json:"path"`
	ID           string          `json:"id"`
	Address      string          `json:"address"`
	Name         string          `json:"name"`
	Alias        string          `json:"alias"`
	Powered      bool            `json:"powered"`
	Discoverable bool            `json:"discoverable"`
	Pairable     bool            `json:"pairable"`
	Discovering  bool            `json:"discovering"`
	UUIDs        []string        `json:"uuids,omitempty"`
}

// NewAdapterFromProps constructs an Adapter from D-Bus object path and properties.
func NewAdapterFromProps(objPath dbus.ObjectPath, props map[string]dbus.Variant) *Adapter {
	a := &Adapter{
		Path: objPath,
		ID:   path.Base(string(objPath)),
	}

	if v, ok := props["Address"]; ok {
		if s, ok := v.Value().(string); ok {
			a.Address = s
		}
	}
	if v, ok := props["Name"]; ok {
		if s, ok := v.Value().(string); ok {
			a.Name = s
		}
	}
	if v, ok := props["Alias"]; ok {
		if s, ok := v.Value().(string); ok {
			a.Alias = s
		}
	}
	if v, ok := props["Powered"]; ok {
		if b, ok := v.Value().(bool); ok {
			a.Powered = b
		}
	}
	if v, ok := props["Discoverable"]; ok {
		if b, ok := v.Value().(bool); ok {
			a.Discoverable = b
		}
	}
	if v, ok := props["Pairable"]; ok {
		if b, ok := v.Value().(bool); ok {
			a.Pairable = b
		}
	}
	if v, ok := props["Discovering"]; ok {
		if b, ok := v.Value().(bool); ok {
			a.Discovering = b
		}
	}
	if v, ok := props["UUIDs"]; ok {
		if uuids, ok := v.Value().([]string); ok {
			a.UUIDs = uuids
		}
	}

	return a
}

// DisplayName returns the adapter alias or name.
func (a *Adapter) DisplayName() string {
	if a.Alias != "" {
		return a.Alias
	}
	if a.Name != "" {
		return a.Name
	}
	return a.ID
}
