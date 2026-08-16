package bluez

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	BlueZService           = "org.bluez"
	AdapterInterface       = "org.bluez.Adapter1"
	DeviceInterface        = "org.bluez.Device1"
	PropertiesInterface    = "org.freedesktop.DBus.Properties"
	ObjectManagerInterface = "org.freedesktop.DBus.ObjectManager"
)

// Client defines the interface for interacting with BlueZ over D-Bus.
type Client interface {
	Close() error
	GetAdapters() ([]*Adapter, error)
	GetDefaultAdapter(preferredName string) (*Adapter, error)
	SetAdapterAlias(adapterPath dbus.ObjectPath, alias string) error
	GetDevices(adapterPath dbus.ObjectPath) ([]*Device, error)
	GetDevice(devicePath dbus.ObjectPath) (*Device, error)
	StartDiscovery(adapterPath dbus.ObjectPath) error
	StopDiscovery(adapterPath dbus.ObjectPath) error
	Scan(adapterPath dbus.ObjectPath, timeout time.Duration, onFound func(d *Device)) ([]*Device, error)
	Pair(devicePath dbus.ObjectPath) error
	CancelPairing(devicePath dbus.ObjectPath) error
	Connect(devicePath dbus.ObjectPath) error
	Disconnect(devicePath dbus.ObjectPath) error
	SetTrusted(devicePath dbus.ObjectPath, trusted bool) error
	RemoveDevice(adapterPath, devicePath dbus.ObjectPath) error
}

// DBusClient is the concrete implementation of Client using godbus.
type DBusClient struct {
	conn *dbus.Conn
}

// NewClient establishes a connection to the System D-Bus and verifies BlueZ is reachable.
func NewClient() (*DBusClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to system D-Bus: %w", err)
	}
	return &DBusClient{conn: conn}, nil
}

// NewClientWithConn creates a DBusClient wrapping an existing D-Bus connection.
func NewClientWithConn(conn *dbus.Conn) *DBusClient {
	return &DBusClient{conn: conn}
}

// Close closes the underlying D-Bus connection.
func (c *DBusClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetManagedObjects calls org.freedesktop.DBus.ObjectManager.GetManagedObjects on org.bluez.
func (c *DBusClient) GetManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, error) {
	obj := c.conn.Object(BlueZService, "/")
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call(ObjectManagerInterface+".GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return nil, fmt.Errorf("failed to query BlueZ objects (is bluetooth.service running?): %w", err)
	}
	return objects, nil
}

// GetAdapters returns all Bluetooth adapters managed by BlueZ.
func (c *DBusClient) GetAdapters() ([]*Adapter, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	var adapters []*Adapter
	for objPath, ifaces := range objects {
		if props, ok := ifaces[AdapterInterface]; ok {
			adapters = append(adapters, NewAdapterFromProps(objPath, props))
		}
	}
	return adapters, nil
}

// GetDefaultAdapter finds an adapter by ID or returns the first available adapter.
func (c *DBusClient) GetDefaultAdapter(preferredName string) (*Adapter, error) {
	adapters, err := c.GetAdapters()
	if err != nil {
		return nil, err
	}
	if len(adapters) == 0 {
		return nil, fmt.Errorf("no Bluetooth adapter found (check rfkill and hardware)")
	}

	if preferredName != "" {
		for _, a := range adapters {
			if strings.EqualFold(a.ID, preferredName) || strings.EqualFold(a.Name, preferredName) {
				return a, nil
			}
		}
		return nil, fmt.Errorf("adapter '%s' not found", preferredName)
	}

	return adapters[0], nil
}

// SetAdapterAlias sets the broadcast alias name on the adapter.
func (c *DBusClient) SetAdapterAlias(adapterPath dbus.ObjectPath, alias string) error {
	obj := c.conn.Object(BlueZService, adapterPath)
	err := obj.SetProperty(AdapterInterface+".Alias", dbus.MakeVariant(alias))
	if err != nil {
		return fmt.Errorf("failed to set adapter alias: %w", err)
	}
	return nil
}

// GetDevices returns all devices known to BlueZ for a given adapter.
func (c *DBusClient) GetDevices(adapterPath dbus.ObjectPath) ([]*Device, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	var devices []*Device
	for objPath, ifaces := range objects {
		if props, ok := ifaces[DeviceInterface]; ok {
			dev := NewDeviceFromProps(objPath, props)
			if adapterPath == "" || dev.AdapterPath == adapterPath || strings.HasPrefix(string(objPath), string(adapterPath)+"/") {
				devices = append(devices, dev)
			}
		}
	}
	return devices, nil
}

// GetDevice fetches a single device by its D-Bus object path.
func (c *DBusClient) GetDevice(devicePath dbus.ObjectPath) (*Device, error) {
	obj := c.conn.Object(BlueZService, devicePath)
	var props map[string]dbus.Variant
	err := obj.Call(PropertiesInterface+".GetAll", 0, DeviceInterface).Store(&props)
	if err != nil {
		return nil, fmt.Errorf("device '%s' not found: %w", devicePath, err)
	}
	return NewDeviceFromProps(devicePath, props), nil
}

// StartDiscovery begins active scanning on the adapter.
func (c *DBusClient) StartDiscovery(adapterPath dbus.ObjectPath) error {
	obj := c.conn.Object(BlueZService, adapterPath)
	call := obj.Call(AdapterInterface+".StartDiscovery", 0)
	if call.Err != nil {
		if strings.Contains(call.Err.Error(), "InProgress") {
			return nil
		}
		return fmt.Errorf("failed to start discovery: %w", call.Err)
	}
	return nil
}

// StopDiscovery stops scanning on the adapter.
func (c *DBusClient) StopDiscovery(adapterPath dbus.ObjectPath) error {
	obj := c.conn.Object(BlueZService, adapterPath)
	call := obj.Call(AdapterInterface+".StopDiscovery", 0)
	if call.Err != nil {
		if strings.Contains(call.Err.Error(), "InProgress") || strings.Contains(call.Err.Error(), "NotReady") {
			return nil
		}
		return fmt.Errorf("failed to stop discovery: %w", call.Err)
	}
	return nil
}

// Scan performs discovery for a given duration while emitting found devices.
func (c *DBusClient) Scan(adapterPath dbus.ObjectPath, timeout time.Duration, onFound func(d *Device)) ([]*Device, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// Register match for ObjectManager and Properties signals
	matchRule := "type='signal',sender='org.bluez'"
	c.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)
	defer func() {
		c.conn.BusObject().Call("org.freedesktop.DBus.RemoveMatch", 0, matchRule)
	}()

	signals := make(chan *dbus.Signal, 64)
	c.conn.Signal(signals)
	defer c.conn.RemoveSignal(signals)

	if err := c.StartDiscovery(adapterPath); err != nil {
		return nil, err
	}
	defer func() {
		_ = c.StopDiscovery(adapterPath)
	}()

	deviceMap := make(map[dbus.ObjectPath]*Device)
	var mu sync.Mutex

	// Seed existing devices
	if initialDevs, err := c.GetDevices(adapterPath); err == nil {
		for _, dev := range initialDevs {
			deviceMap[dev.Path] = dev
			if onFound != nil {
				onFound(dev)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := false
	for !done {
		select {
		case <-ctx.Done():
			done = true
		case sig, ok := <-signals:
			if !ok {
				done = true
				break
			}
			if sig == nil {
				continue
			}

			switch sig.Name {
			case ObjectManagerInterface + ".InterfacesAdded":
				if len(sig.Body) >= 2 {
					objPath, ok1 := sig.Body[0].(dbus.ObjectPath)
					ifaces, ok2 := sig.Body[1].(map[string]map[string]dbus.Variant)
					if ok1 && ok2 {
						if props, ok := ifaces[DeviceInterface]; ok {
							mu.Lock()
							dev := NewDeviceFromProps(objPath, props)
							deviceMap[objPath] = dev
							mu.Unlock()
							if onFound != nil {
								onFound(dev)
							}
						}
					}
				}

			case PropertiesInterface + ".PropertiesChanged":
				if len(sig.Body) >= 2 {
					ifaceName, ok1 := sig.Body[0].(string)
					changedProps, ok2 := sig.Body[1].(map[string]dbus.Variant)
					if ok1 && ok2 && ifaceName == DeviceInterface {
						mu.Lock()
						dev, exists := deviceMap[sig.Path]
						if exists {
							// Update device properties
							if v, ok := changedProps["RSSI"]; ok {
								if n, ok := v.Value().(int16); ok {
									dev.RSSI = n
								}
							}
							if v, ok := changedProps["Name"]; ok {
								if s, ok := v.Value().(string); ok {
									dev.Name = s
								}
							}
							if v, ok := changedProps["Alias"]; ok {
								if s, ok := v.Value().(string); ok {
									dev.Alias = s
								}
							}
							if v, ok := changedProps["Connected"]; ok {
								if b, ok := v.Value().(bool); ok {
									dev.Connected = b
								}
							}
							if v, ok := changedProps["Paired"]; ok {
								if b, ok := v.Value().(bool); ok {
									dev.Paired = b
								}
							}
							if v, ok := changedProps["UUIDs"]; ok {
								if uuids, ok := v.Value().([]string); ok {
									dev.UUIDs = uuids
									dev.AudioCaps = DetectAudioCapabilities(uuids)
								}
							}
						}
						mu.Unlock()
						if exists && onFound != nil {
							onFound(dev)
						}
					}
				}
			}
		}
	}

	mu.Lock()
	defer mu.Unlock()
	var result []*Device
	for _, d := range deviceMap {
		result = append(result, d)
	}
	return result, nil
}

// Pair initiates pairing with the device.
func (c *DBusClient) Pair(devicePath dbus.ObjectPath) error {
	agent := NewAgent(c.conn, "KeyboardDisplay")
	_ = agent.Register()
	defer agent.Unregister()

	obj := c.conn.Object(BlueZService, devicePath)
	call := obj.Call(DeviceInterface+".Pair", 0)
	if call.Err != nil {
		errMsg := call.Err.Error()
		if strings.Contains(errMsg, "AlreadyExists") || strings.Contains(errMsg, "AlreadyPaired") {
			return nil
		}
		return fmt.Errorf("pairing failed: %w", call.Err)
	}
	return nil
}

// CancelPairing cancels an ongoing pairing attempt.
func (c *DBusClient) CancelPairing(devicePath dbus.ObjectPath) error {
	obj := c.conn.Object(BlueZService, devicePath)
	return obj.Call(DeviceInterface+".CancelPairing", 0).Err
}

// Connect connects to the device.
func (c *DBusClient) Connect(devicePath dbus.ObjectPath) error {
	obj := c.conn.Object(BlueZService, devicePath)
	call := obj.Call(DeviceInterface+".Connect", 0)
	if call.Err != nil {
		errMsg := call.Err.Error()
		if strings.Contains(errMsg, "AlreadyConnected") {
			return nil
		}
		if strings.Contains(errMsg, "InProgress") {
			return nil
		}
		return fmt.Errorf("connection failed: %w", call.Err)
	}
	return nil
}

// Disconnect disconnects from the device.
func (c *DBusClient) Disconnect(devicePath dbus.ObjectPath) error {
	obj := c.conn.Object(BlueZService, devicePath)
	call := obj.Call(DeviceInterface+".Disconnect", 0)
	if call.Err != nil {
		errMsg := call.Err.Error()
		if strings.Contains(errMsg, "NotConnected") {
			return nil
		}
		return fmt.Errorf("disconnect failed: %w", call.Err)
	}
	return nil
}

// SetTrusted sets the Trusted property on the device.
func (c *DBusClient) SetTrusted(devicePath dbus.ObjectPath, trusted bool) error {
	obj := c.conn.Object(BlueZService, devicePath)
	call := obj.Call(PropertiesInterface+".Set", 0, DeviceInterface, "Trusted", dbus.MakeVariant(trusted))
	if call.Err != nil {
		return fmt.Errorf("failed to set trusted: %w", call.Err)
	}
	return nil
}

// RemoveDevice removes a device from BlueZ storage.
func (c *DBusClient) RemoveDevice(adapterPath, devicePath dbus.ObjectPath) error {
	obj := c.conn.Object(BlueZService, adapterPath)
	return obj.Call(AdapterInterface+".RemoveDevice", 0, devicePath).Err
}

// DevicePathFromAddress converts a MAC address and adapter path to a BlueZ D-Bus object path.
func DevicePathFromAddress(adapterPath dbus.ObjectPath, address string) dbus.ObjectPath {
	addrClean := strings.ToUpper(strings.ReplaceAll(address, ":", "_"))
	return dbus.ObjectPath(path.Join(string(adapterPath), "dev_"+addrClean))
}
