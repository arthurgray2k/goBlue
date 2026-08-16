package bluez

// AudioCapabilities describes the Bluetooth audio profile roles exposed by a device.
type AudioCapabilities struct {
	// A2DPSink indicates the device can receive high-quality audio streams (headphones, speakers, TVs acting as audio outputs).
	HasA2DPSink bool `json:"a2dp_sink"`
	// A2DPSource indicates the device can transmit high-quality audio streams (phones, PCs).
	HasA2DPSource bool `json:"a2dp_source"`
	// HasAVRCP indicates Audio/Video Remote Control capability (play/pause/volume).
	HasAVRCP bool `json:"avrcp"`
	// HasHFP indicates Handsfree Profile (voice calls / microphone).
	HasHFP bool `json:"hfp"`
	// HasHSP indicates Headset Profile (legacy voice telephony).
	HasHSP bool `json:"hsp"`
}

// CanReceiveAudio returns true if the device supports receiving audio output from this Linux machine.
func (a AudioCapabilities) CanReceiveAudio() bool {
	return a.HasA2DPSink || a.HasHFP || a.HasHSP
}

// DetectAudioCapabilities inspects the device UUIDs and returns its audio roles.
// It never infers capabilities from device names or icons.
func DetectAudioCapabilities(uuids []string) AudioCapabilities {
	var caps AudioCapabilities
	for _, raw := range uuids {
		norm := NormalizeUUID(raw)
		switch norm {
		case UUIDA2DPSink:
			caps.HasA2DPSink = true
		case UUIDA2DPSource:
			caps.HasA2DPSource = true
		case UUIDAVRCPTarget, UUIDAVRCPController:
			caps.HasAVRCP = true
		case UUIDHandsfree, UUIDHandsfreeAudioGateway:
			caps.HasHFP = true
		case UUIDHeadset, UUIDHeadsetAudioGateway:
			caps.HasHSP = true
		}
	}
	return caps
}
