package bluez

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	AgentPath      = "/org/bluez/goBlue/agent"
	AgentInterface = "org.bluez.Agent1"
)

// Agent implements the BlueZ org.bluez.Agent1 D-Bus interface.
type Agent struct {
	conn       *dbus.Conn
	capability string
	registered bool
}

// NewAgent creates a new Agent instance.
func NewAgent(conn *dbus.Conn, capability string) *Agent {
	if capability == "" {
		capability = "KeyboardDisplay"
	}
	return &Agent{
		conn:       conn,
		capability: capability,
	}
}

// Register registers the agent with BlueZ AgentManager1.
func (a *Agent) Register() error {
	if a.conn == nil {
		return fmt.Errorf("no D-Bus connection")
	}

	err := a.conn.Export(a, AgentPath, AgentInterface)
	if err != nil {
		return fmt.Errorf("failed to export agent: %w", err)
	}

	agentManager := a.conn.Object(BlueZService, dbus.ObjectPath("/org/bluez"))
	call := agentManager.Call("org.bluez.AgentManager1.RegisterAgent", 0, dbus.ObjectPath(AgentPath), a.capability)
	if call.Err != nil {
		return fmt.Errorf("failed to register agent with BlueZ: %w", call.Err)
	}

	_ = agentManager.Call("org.bluez.AgentManager1.RequestDefaultAgent", 0, dbus.ObjectPath(AgentPath))
	a.registered = true
	return nil
}

// Unregister unregisters the agent from BlueZ.
func (a *Agent) Unregister() {
	if !a.registered || a.conn == nil {
		return
	}
	agentManager := a.conn.Object(BlueZService, dbus.ObjectPath("/org/bluez"))
	_ = agentManager.Call("org.bluez.AgentManager1.UnregisterAgent", 0, dbus.ObjectPath(AgentPath))
	_ = a.conn.Export(nil, AgentPath, AgentInterface)
	a.registered = false
}

// Release is called when BlueZ unregisters the agent.
func (a *Agent) Release() *dbus.Error {
	return nil
}

// RequestPinCode requests a PIN code for pairing.
func (a *Agent) RequestPinCode(device dbus.ObjectPath) (string, *dbus.Error) {
	fmt.Printf("\nEnter PIN code for pairing: ")
	reader := bufio.NewReader(os.Stdin)
	pin, err := reader.ReadString('\n')
	if err != nil {
		return "", dbus.NewError("org.bluez.Error.Canceled", nil)
	}
	return strings.TrimSpace(pin), nil
}

// DisplayPinCode displays a PIN code to the user.
func (a *Agent) DisplayPinCode(device dbus.ObjectPath, pincode string) *dbus.Error {
	fmt.Printf("\nPairing PIN code: %s\n", pincode)
	return nil
}

// RequestPasskey requests a numeric passkey (0-999999).
func (a *Agent) RequestPasskey(device dbus.ObjectPath) (uint32, *dbus.Error) {
	fmt.Printf("\nEnter passkey (0-999999): ")
	var passkey uint32
	_, err := fmt.Scanf("%d\n", &passkey)
	if err != nil {
		return 0, dbus.NewError("org.bluez.Error.Canceled", nil)
	}
	return passkey, nil
}

// DisplayPasskey displays passkey and entered digits.
func (a *Agent) DisplayPasskey(device dbus.ObjectPath, passkey uint32, entered uint16) *dbus.Error {
	fmt.Printf("\nPasskey: %06d (entered: %d)\n", passkey, entered)
	return nil
}

// RequestConfirmation asks the user to confirm a passkey.
func (a *Agent) RequestConfirmation(device dbus.ObjectPath, passkey uint32) *dbus.Error {
	fmt.Printf("\nConfirm passkey %06d (yes/no): ", passkey)
	reader := bufio.NewReader(os.Stdin)
	ans, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(ans)), "y") {
		return dbus.NewError("org.bluez.Error.Rejected", nil)
	}
	return nil
}

// RequestAuthorization asks the user to authorize a connection.
func (a *Agent) RequestAuthorization(device dbus.ObjectPath) *dbus.Error {
	return nil
}

// AuthorizeService authorizes a specific service UUID.
func (a *Agent) AuthorizeService(device dbus.ObjectPath, uuid string) *dbus.Error {
	return nil
}

// Cancel is called when pairing is canceled by BlueZ.
func (a *Agent) Cancel() *dbus.Error {
	return nil
}
