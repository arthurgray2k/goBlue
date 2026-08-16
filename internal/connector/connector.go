package connector

import (
	"fmt"
	"time"

	"github.com/arthurgray2k/goBlue/internal/bluez"
	"github.com/arthurgray2k/goBlue/internal/resolver"
	"github.com/godbus/dbus/v5"
)

// ConnectResult holds the outcome and capability details of a connection attempt.
type ConnectResult struct {
	Device           *bluez.Device `json:"device"`
	AlreadyConnected bool          `json:"already_connected"`
	HasA2DPSink      bool          `json:"has_a2dp_sink"`
	HasAudioCap      bool          `json:"has_audio_capability"`
}

// DisconnectResult holds the outcome of a disconnection attempt.
type DisconnectResult struct {
	Device              *bluez.Device `json:"device"`
	AlreadyDisconnected bool          `json:"already_disconnected"`
}

// Connector handles device connection and disconnection workflows.
type Connector struct {
	client bluez.Client
}

// NewConnector creates a new Connector wrapping the BlueZ client.
func NewConnector(client bluez.Client) *Connector {
	return &Connector{client: client}
}

// Connect resolves a device, evaluates audio roles, and establishes the Bluetooth connection.
func (c *Connector) Connect(adapterPath dbus.ObjectPath, identifier string) (*ConnectResult, error) {
	dev, err := resolver.ResolveDevice(c.client, adapterPath, identifier)
	if err != nil {
		return nil, err
	}

	result := &ConnectResult{
		Device:           dev,
		AlreadyConnected: dev.Connected,
		HasA2DPSink:      dev.AudioCaps.HasA2DPSink,
		HasAudioCap:      dev.AudioCaps.CanReceiveAudio(),
	}

	if dev.Connected {
		return result, nil
	}

	if err := c.client.Connect(dev.Path); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return result, nil
}

// Disconnect resolves a device and closes its active connection.
func (c *Connector) Disconnect(adapterPath dbus.ObjectPath, identifier string) (*DisconnectResult, error) {
	dev, err := resolver.ResolveDevice(c.client, adapterPath, identifier)
	if err != nil {
		return nil, err
	}

	result := &DisconnectResult{
		Device:              dev,
		AlreadyDisconnected: !dev.Connected,
	}

	if !dev.Connected {
		return result, nil
	}

	if err := c.client.Disconnect(dev.Path); err != nil {
		return nil, fmt.Errorf("disconnect failed: %w", err)
	}

	return result, nil
}

// EnsureConnected verifies that a device is connected, connecting it if currently disconnected.
func (c *Connector) EnsureConnected(adapterPath dbus.ObjectPath, identifier string) (*bluez.Device, error) {
	dev, err := resolver.ResolveDevice(c.client, adapterPath, identifier)
	if err != nil {
		return nil, err
	}

	if !dev.Connected {
		if err := c.client.Connect(dev.Path); err != nil {
			return nil, fmt.Errorf("connection failed: %w", err)
		}
		// Brief pause for PipeWire / BlueZ subsystem to populate node
		time.Sleep(1 * time.Second)
		dev.Connected = true
	}

	return dev, nil
}
