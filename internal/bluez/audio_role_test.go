package bluez

import (
	"testing"
)

func TestDetectAudioCapabilities(t *testing.T) {
	t.Run("Headset with A2DP Sink, AVRCP, and HFP", func(t *testing.T) {
		uuids := []string{
			"0000110b-0000-1000-8000-00805f9b34fb", // A2DP Sink
			"0000110c-0000-1000-8000-00805f9b34fb", // AVRCP Target
			"0000110e-0000-1000-8000-00805f9b34fb", // AVRCP Controller
			"0000111e-0000-1000-8000-00805f9b34fb", // Handsfree HF
		}
		caps := DetectAudioCapabilities(uuids)
		if !caps.HasA2DPSink {
			t.Errorf("expected HasA2DPSink to be true")
		}
		if caps.HasA2DPSource {
			t.Errorf("expected HasA2DPSource to be false")
		}
		if !caps.HasAVRCP {
			t.Errorf("expected HasAVRCP to be true")
		}
		if !caps.HasHFP {
			t.Errorf("expected HasHFP to be true")
		}
		if !caps.CanReceiveAudio() {
			t.Errorf("expected CanReceiveAudio() to be true")
		}
	})

	t.Run("Audio Source device (Smartphone/Transmitter)", func(t *testing.T) {
		uuids := []string{
			"0000110a-0000-1000-8000-00805f9b34fb", // A2DP Source
			"0000111f-0000-1000-8000-00805f9b34fb", // HFP AG
		}
		caps := DetectAudioCapabilities(uuids)
		if caps.HasA2DPSink {
			t.Errorf("expected HasA2DPSink to be false")
		}
		if !caps.HasA2DPSource {
			t.Errorf("expected HasA2DPSource to be true")
		}
		if !caps.HasHFP {
			t.Errorf("expected HasHFP to be true (AG)")
		}
	})

	t.Run("Non-audio device (Keyboard/Mouse)", func(t *testing.T) {
		uuids := []string{
			"00001124-0000-1000-8000-00805f9b34fb", // HID
			"00001800-0000-1000-8000-00805f9b34fb", // GAP
		}
		caps := DetectAudioCapabilities(uuids)
		if caps.HasA2DPSink || caps.HasA2DPSource || caps.HasAVRCP || caps.HasHFP || caps.HasHSP {
			t.Errorf("expected no audio capabilities, got %+v", caps)
		}
		if caps.CanReceiveAudio() {
			t.Errorf("expected CanReceiveAudio() to be false")
		}
	})

	t.Run("Legacy Headset profile (HSP)", func(t *testing.T) {
		uuids := []string{
			"00001108-0000-1000-8000-00805f9b34fb", // HSP
		}
		caps := DetectAudioCapabilities(uuids)
		if !caps.HasHSP {
			t.Errorf("expected HasHSP to be true")
		}
		if !caps.CanReceiveAudio() {
			t.Errorf("expected CanReceiveAudio() to be true for HSP")
		}
	})
}
