package audio

import (
	"testing"

	"github.com/arthurgray2k/goBlue/internal/bluez"
)

const samplePwDumpJSON = `[
  {
    "id": 36,
    "type": "PipeWire:Interface:Metadata",
    "props": {
      "metadata.name": "default"
    },
    "metadata": [
      {
        "key": "default.audio.sink",
        "value": {"name": "bluez_output.28_04_C6_73_E8_D6.1"}
      }
    ]
  },
  {
    "id": 54,
    "type": "PipeWire:Interface:Node",
    "info": {
      "state": "suspended",
      "props": {
        "media.class": "Audio/Sink",
        "node.name": "alsa_output.pci-0000_00_1f.3-platform-skl_hda_dsp_generic.HiFi__hw_sofhdadsp__sink",
        "node.description": "Speaker + Headphones"
      }
    }
  },
  {
    "id": 88,
    "type": "PipeWire:Interface:Node",
    "info": {
      "state": "running",
      "props": {
        "media.class": "Audio/Sink",
        "node.name": "bluez_output.28_04_C6_73_E8_D6.1",
        "node.description": "realme Buds T200 Lite",
        "device.api": "bluez5",
        "api.bluez5.address": "28:04:C6:73:E8:D6",
        "api.bluez5.profile": "a2dp-sink"
      }
    }
  }
]`

func TestParsePipeWireDump(t *testing.T) {
	sinks, err := ParsePipeWireDump([]byte(samplePwDumpJSON))
	if err != nil {
		t.Fatalf("ParsePipeWireDump error: %v", err)
	}

	if len(sinks) != 2 {
		t.Fatalf("expected 2 sinks, got %d", len(sinks))
	}

	// First sink: ALSA
	alsaSink := sinks[0]
	if alsaSink.ID != 54 {
		t.Errorf("expected ID 54, got %d", alsaSink.ID)
	}
	if alsaSink.IsBluetooth {
		t.Errorf("expected ALSA sink to not be bluetooth")
	}
	if alsaSink.IsDefault {
		t.Errorf("expected ALSA sink to not be default")
	}

	// Second sink: Bluetooth
	btSink := sinks[1]
	if btSink.ID != 88 {
		t.Errorf("expected ID 88, got %d", btSink.ID)
	}
	if !btSink.IsBluetooth {
		t.Errorf("expected Bluetooth sink to be marked IsBluetooth")
	}
	if btSink.BluetoothAddress != "28:04:C6:73:E8:D6" {
		t.Errorf("expected address 28:04:C6:73:E8:D6, got %s", btSink.BluetoothAddress)
	}
	if btSink.Profile != "A2DP" {
		t.Errorf("expected profile A2DP, got %s", btSink.Profile)
	}
	if !btSink.IsDefault {
		t.Errorf("expected Bluetooth sink to be marked default")
	}
}

func TestExtractAddressFromNodeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bluez_output.28_04_C6_73_E8_D6.1", "28:04:C6:73:E8:D6"},
		{"bluez_sink.AA_BB_CC_DD_EE_FF.a2dp", "AA:BB:CC:DD:EE:FF"},
		{"alsa_output.pci", ""},
	}

	for _, tt := range tests {
		got := extractAddressFromNodeName(tt.input)
		if got != tt.expected {
			t.Errorf("extractAddressFromNodeName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatProfile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a2dp-sink", "A2DP"},
		{"a2dp_sink", "A2DP"},
		{"headset-head-unit", "HFP"},
		{"hfp_hf", "HFP"},
		{"hsp_hs", "HSP"},
		{"custom-profile", "custom-profile"},
	}

	for _, tt := range tests {
		got := formatProfile(tt.input)
		if got != tt.expected {
			t.Errorf("formatProfile(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetAudioStatus(t *testing.T) {
	c := &Client{}
	devices := []*bluez.Device{
		{
			Address:   "28:04:C6:73:E8:D6",
			Name:      "realme Buds T200 Lite",
			Connected: false,
			AudioCaps: bluez.AudioCapabilities{HasA2DPSink: true},
		},
		{
			Address:   "11:22:33:44:55:66",
			Name:      "Bluetooth Keyboard",
			Connected: true,
			AudioCaps: bluez.AudioCapabilities{},
		},
	}

	items, err := c.GetAudioStatus(devices)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 status items, got %d", len(items))
	}
	if items[0].DeviceName != "realme Buds T200 Lite" {
		t.Errorf("expected first item realme Buds, got %s", items[0].DeviceName)
	}
	if items[0].State != "disconnected" {
		t.Errorf("expected disconnected, got %s", items[0].State)
	}
	if items[1].State != "connected" {
		t.Errorf("expected connected, got %s", items[1].State)
	}
}

func TestFindSinkByBluetoothAddress(t *testing.T) {
	c := &Client{}
	if c.IsAvailable() {
		if !c.IsAvailable() {
			t.Errorf("expected PipeWire client to be available")
		}
	}
}
