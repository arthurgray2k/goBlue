package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthurgray2k/goBlue/internal/audio"
	"github.com/arthurgray2k/goBlue/internal/bluez"
)

func TestRenderAdapter(t *testing.T) {
	var buf bytes.Buffer
	adapter := &bluez.Adapter{
		ID:           "hci0",
		Address:      "38:7A:0E:12:20:7C",
		Name:         "mint-external",
		Alias:        "My Custom Laptop",
		Powered:      true,
		Discoverable: false,
		Pairable:     true,
		Discovering:  false,
	}

	RenderAdapter(&buf, adapter)
	out := buf.String()

	if !strings.Contains(out, "BLUETOOTH ADAPTER") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "hci0") || !strings.Contains(out, "mint-external") || !strings.Contains(out, "My Custom Laptop") {
		t.Errorf("missing adapter details: %s", out)
	}
}

func TestRenderDevices(t *testing.T) {
	var buf bytes.Buffer
	devices := []*bluez.Device{
		{
			Address:   "AA:BB:CC:DD:EE:FF",
			Name:      "Sony WH-1000XM5",
			Paired:    true,
			Trusted:   true,
			Connected: true,
		},
		{
			Address:   "11:22:33:44:55:66",
			Name:      "Galaxy Buds",
			Paired:    true,
			Trusted:   true,
			Connected: false,
		},
	}

	RenderDevices(&buf, devices)
	out := buf.String()

	if !strings.Contains(out, "KNOWN BLUETOOTH DEVICES") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "Sony WH-1000XM5") || !strings.Contains(out, "Galaxy Buds") {
		t.Errorf("missing device entries: %s", out)
	}
}

func TestRenderScanResults(t *testing.T) {
	var buf bytes.Buffer
	devices := []*bluez.Device{
		{
			Address: "AA:BB:CC:DD:EE:FF",
			Name:    "Sony WH-1000XM5",
			RSSI:    -48,
			Paired:  true,
		},
	}

	RenderScanResults(&buf, devices)
	out := buf.String()

	if !strings.Contains(out, "BLUETOOTH DEVICES") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "-48") {
		t.Errorf("missing RSSI: %s", out)
	}
}

func TestRenderStatus(t *testing.T) {
	var buf bytes.Buffer
	adapter := &bluez.Adapter{ID: "hci0", Powered: true}
	dev := &bluez.Device{
		Address:   "AA:BB:CC:DD:EE:FF",
		Name:      "Sony WH-1000XM5",
		Paired:    true,
		Trusted:   true,
		Connected: true,
	}
	extras := []*DeviceStatusExtra{
		{
			Device:      dev,
			HasAudioCap: true,
			Profile:     "A2DP",
			IsDefault:   true,
			AudioSink:   &audio.AudioSink{ID: 88, Name: "bluez_output"},
		},
	}

	RenderStatus(&buf, adapter, extras)
	out := buf.String()

	if !strings.Contains(out, "BLUETOOTH STATUS") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "Sony WH-1000XM5") || !strings.Contains(out, "A2DP") {
		t.Errorf("missing status info: %s", out)
	}
}

func TestRenderDeviceInfo(t *testing.T) {
	var buf bytes.Buffer
	dev := &bluez.Device{
		Address:   "AA:BB:CC:DD:EE:FF",
		Name:      "Living Room TV",
		Paired:    false,
		Trusted:   false,
		Connected: false,
		AudioCaps: bluez.AudioCapabilities{
			HasA2DPSink:   true,
			HasA2DPSource: false,
		},
	}

	RenderDeviceInfo(&buf, dev)
	out := buf.String()

	if !strings.Contains(out, "DEVICE") || !strings.Contains(out, "AUDIO CAPABILITIES") {
		t.Errorf("missing headers: %s", out)
	}
	if !strings.Contains(out, "A2DP Sink:") || !strings.Contains(out, "yes") {
		t.Errorf("expected A2DP Sink yes: %s", out)
	}
}

func TestRenderAudioStatus(t *testing.T) {
	var buf bytes.Buffer
	items := []*audio.AudioStatusItem{
		{
			DeviceName: "Sony WH-1000XM5",
			Address:    "AA:BB:CC:DD:EE:FF",
			State:      "connected",
			Profile:    "A2DP",
			IsDefault:  true,
		},
	}

	RenderAudioStatus(&buf, items)
	out := buf.String()

	if !strings.Contains(out, "BLUETOOTH AUDIO") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "Sony WH-1000XM5") || !strings.Contains(out, "A2DP") {
		t.Errorf("missing audio items: %s", out)
	}
}

func TestRenderAudioInfo(t *testing.T) {
	var buf bytes.Buffer
	dev := &bluez.Device{
		Address: "AA:BB:CC:DD:EE:FF",
		Name:    "Living Room TV",
		AudioCaps: bluez.AudioCapabilities{
			HasA2DPSink: true,
		},
	}

	RenderAudioInfo(&buf, dev, true)
	out := buf.String()

	if !strings.Contains(out, "Can receive audio from this Linux machine:") || !strings.Contains(out, "yes") {
		t.Errorf("missing can receive audio string: %s", out)
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	err := RenderJSON(&buf, data)
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}
	if !strings.Contains(buf.String(), `"key": "value"`) {
		t.Errorf("unexpected json: %s", buf.String())
	}
}
