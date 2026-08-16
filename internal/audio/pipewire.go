package audio

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/arthurgray2k/goBlue/internal/bluez"
)

// PipeWireObject represents an object from pw-dump JSON.
type PipeWireObject struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Props struct {
		MetadataName string `json:"metadata.name"`
	} `json:"props"`
	Info struct {
		Props map[string]interface{} `json:"props"`
		State string                 `json:"state"`
	} `json:"info"`
	Metadata []struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	} `json:"metadata"`
}

// Client provides PipeWire/WirePlumber audio inspection and management.
type Client struct{}

// NewClient creates a new PipeWire audio client.
func NewClient() *Client {
	return &Client{}
}

// IsAvailable checks if PipeWire tools (pw-dump or wpctl) are accessible.
func (c *Client) IsAvailable() bool {
	_, err := exec.LookPath("pw-dump")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("wpctl")
	return err == nil
}

// GetSinks returns all Audio/Sink nodes from PipeWire.
func (c *Client) GetSinks() ([]*AudioSink, error) {
	pwDumpPath, err := exec.LookPath("pw-dump")
	if err != nil {
		return nil, fmt.Errorf("pw-dump not found: %w", err)
	}

	cmd := exec.Command(pwDumpPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run pw-dump: %w", err)
	}

	return ParsePipeWireDump(output)
}

// ParsePipeWireDump parses pw-dump JSON bytes into AudioSink structs.
func ParsePipeWireDump(data []byte) ([]*AudioSink, error) {
	var objects []PipeWireObject
	if err := json.Unmarshal(data, &objects); err != nil {
		return nil, fmt.Errorf("failed to parse pw-dump JSON: %w", err)
	}

	// 1. Find default sink name from metadata
	var defaultSinkName string
	for _, obj := range objects {
		if obj.Type == "PipeWire:Interface:Metadata" && obj.Props.MetadataName == "default" {
			for _, m := range obj.Metadata {
				if m.Key == "default.audio.sink" || (defaultSinkName == "" && m.Key == "default.configured.audio.sink") {
					var valObj struct {
						Name string `json:"name"`
					}
					if err := json.Unmarshal(m.Value, &valObj); err == nil && valObj.Name != "" {
						defaultSinkName = valObj.Name
					} else {
						var strVal string
						if err := json.Unmarshal(m.Value, &strVal); err == nil && strVal != "" {
							defaultSinkName = strVal
						}
					}
				}
			}
		}
	}

	// 2. Find Audio/Sink nodes
	var sinks []*AudioSink
	for _, obj := range objects {
		if obj.Type != "PipeWire:Interface:Node" {
			continue
		}
		props := obj.Info.Props
		if props == nil {
			continue
		}

		mediaClass, _ := props["media.class"].(string)
		if mediaClass != "Audio/Sink" {
			continue
		}

		sink := &AudioSink{
			ID:    obj.ID,
			State: obj.Info.State,
		}

		if name, ok := props["node.name"].(string); ok {
			sink.Name = name
		}
		if desc, ok := props["node.description"].(string); ok {
			sink.Description = desc
		} else if desc, ok := props["node.nick"].(string); ok {
			sink.Description = desc
		}

		// Detect Bluetooth properties
		deviceAPI, _ := props["device.api"].(string)
		deviceBus, _ := props["device.bus"].(string)
		if deviceAPI == "bluez5" || deviceBus == "bluetooth" || strings.HasPrefix(sink.Name, "bluez_") {
			sink.IsBluetooth = true
		}

		if addr, ok := props["api.bluez5.address"].(string); ok {
			sink.BluetoothAddress = strings.ToUpper(addr)
		} else if sink.IsBluetooth {
			// Extract from node name e.g. bluez_output.28_04_C6_73_E8_D6.1
			sink.BluetoothAddress = extractAddressFromNodeName(sink.Name)
		}

		if prof, ok := props["api.bluez5.profile"].(string); ok {
			sink.Profile = formatProfile(prof)
		} else if strings.Contains(sink.Name, "a2dp") {
			sink.Profile = "A2DP"
		} else if strings.Contains(sink.Name, "headset") || strings.Contains(sink.Name, "hfp") {
			sink.Profile = "HFP"
		}

		if sink.Name != "" && defaultSinkName != "" && sink.Name == defaultSinkName {
			sink.IsDefault = true
		}

		sinks = append(sinks, sink)
	}

	return sinks, nil
}

func formatProfile(raw string) string {
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "a2dp") {
		return "A2DP"
	}
	if strings.Contains(lower, "headset") || strings.Contains(lower, "hfp") {
		return "HFP"
	}
	if strings.Contains(lower, "hsp") {
		return "HSP"
	}
	return raw
}

func extractAddressFromNodeName(name string) string {
	parts := strings.Split(name, ".")
	for _, p := range parts {
		sub := strings.Split(p, "_")
		if len(sub) == 6 {
			return strings.ToUpper(strings.Join(sub, ":"))
		}
	}
	return ""
}

// GetBluetoothSinks returns only Bluetooth audio sinks.
func (c *Client) GetBluetoothSinks() ([]*AudioSink, error) {
	allSinks, err := c.GetSinks()
	if err != nil {
		return nil, err
	}
	var btSinks []*AudioSink
	for _, s := range allSinks {
		if s.IsBluetooth {
			btSinks = append(btSinks, s)
		}
	}
	return btSinks, nil
}

// FindSinkByBluetoothAddress finds a PipeWire sink matching a Bluetooth MAC address.
func (c *Client) FindSinkByBluetoothAddress(address string) (*AudioSink, error) {
	sinks, err := c.GetBluetoothSinks()
	if err != nil {
		return nil, err
	}
	normAddr := strings.ToUpper(strings.ReplaceAll(address, "-", ":"))
	for _, s := range sinks {
		if strings.ToUpper(s.BluetoothAddress) == normAddr {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no PipeWire audio sink found for Bluetooth address %s", address)
}

// SetDefaultSink configures the given sink as the default audio output in PipeWire / WirePlumber.
func (c *Client) SetDefaultSink(sink *AudioSink) error {
	if sink == nil {
		return fmt.Errorf("nil sink provided")
	}

	// 1. Try wpctl set-default <id>
	if wpctlPath, err := exec.LookPath("wpctl"); err == nil {
		cmd := exec.Command(wpctlPath, "set-default", strconv.Itoa(sink.ID))
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			// If failed with ID, try fallback
			_ = out
		}
	}

	// 2. Try pactl set-default-sink <name>
	if pactlPath, err := exec.LookPath("pactl"); err == nil && sink.Name != "" {
		cmd := exec.Command(pactlPath, "set-default-sink", sink.Name)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			return fmt.Errorf("pactl set-default-sink failed: %s (%w)", string(out), err)
		}
	}

	return fmt.Errorf("could not set default sink (neither wpctl nor pactl succeeded)")
}

// GetAudioStatus generates the status list for all known Bluetooth devices and their PipeWire sink state.
func (c *Client) GetAudioStatus(knownDevices []*bluez.Device) ([]*AudioStatusItem, error) {
	var sinks []*AudioSink
	if c.IsAvailable() {
		sinks, _ = c.GetBluetoothSinks()
	}

	sinkByAddr := make(map[string]*AudioSink)
	for _, s := range sinks {
		if s.BluetoothAddress != "" {
			sinkByAddr[s.BluetoothAddress] = s
		}
	}

	var items []*AudioStatusItem
	seenAddrs := make(map[string]bool)

	for _, dev := range knownDevices {
		normAddr := strings.ToUpper(dev.Address)
		seenAddrs[normAddr] = true

		item := &AudioStatusItem{
			DeviceName:  dev.DisplayName(),
			Address:     dev.Address,
			HasAudioCap: dev.AudioCaps.CanReceiveAudio(),
		}

		if dev.Connected {
			item.State = "connected"
		} else {
			item.State = "disconnected"
		}

		if sink, ok := sinkByAddr[normAddr]; ok {
			item.SinkID = sink.ID
			item.SinkName = sink.Name
			item.Profile = sink.Profile
			item.IsDefault = sink.IsDefault
		} else if dev.Connected && dev.AudioCaps.HasA2DPSink {
			item.Profile = "A2DP"
		}

		// Only include devices that are audio capable or already connected as an audio sink
		if item.HasAudioCap || sinkByAddr[normAddr] != nil || dev.Connected {
			items = append(items, item)
		}
	}

	return items, nil
}
