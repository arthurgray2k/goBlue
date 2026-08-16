package audio

// AudioSink represents a Linux audio output sink (PipeWire node / ALSA / PulseAudio).
type AudioSink struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	BluetoothAddress string  `json:"bluetooth_address,omitempty"`
	Profile          string  `json:"profile,omitempty"`
	IsDefault        bool    `json:"is_default"`
	IsBluetooth      bool    `json:"is_bluetooth"`
	State            string  `json:"state,omitempty"`
	Volume           float64 `json:"volume,omitempty"`
	Muted            bool    `json:"muted,omitempty"`
}

// AudioStatusItem represents an entry in the "goBlue audio status" overview.
type AudioStatusItem struct {
	DeviceName  string `json:"device_name"`
	Address     string `json:"address"`
	State       string `json:"state"` // "connected", "disconnected"
	SinkID      int    `json:"sink_id,omitempty"`
	SinkName    string `json:"sink_name,omitempty"`
	Profile     string `json:"profile,omitempty"`
	IsDefault   bool   `json:"is_default"`
	HasAudioCap bool   `json:"has_audio_capability"`
}
