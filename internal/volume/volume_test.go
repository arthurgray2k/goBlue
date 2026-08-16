package volume

import (
	"testing"
)

func TestParseVolumeOutput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedPct int
		expectedMut bool
	}{
		{
			name:        "Standard unmuted volume",
			input:       "Volume: 0.85\n",
			expectedPct: 85,
			expectedMut: false,
		},
		{
			name:        "Muted volume",
			input:       "Volume: 0.50 [MUTED]",
			expectedPct: 50,
			expectedMut: true,
		},
		{
			name:        "100% volume",
			input:       "Volume: 1.00",
			expectedPct: 100,
			expectedMut: false,
		},
		{
			name:        "0% volume",
			input:       "Volume: 0.00",
			expectedPct: 0,
			expectedMut: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseVolumeOutput(tt.input)
			if err != nil {
				t.Fatalf("ParseVolumeOutput error: %v", err)
			}
			if info.Percentage != tt.expectedPct {
				t.Errorf("expected percentage %d, got %d", tt.expectedPct, info.Percentage)
			}
			if info.Muted != tt.expectedMut {
				t.Errorf("expected muted %v, got %v", tt.expectedMut, info.Muted)
			}
		})
	}
}

func TestNormalizeVolumeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+10%", "10%+"},
		{"+5", "5%+"},
		{"-5%", "5%-"},
		{"-10", "10%-"},
		{"80", "80%"},
		{"80%", "80%"},
		{"0.85", "0.85"},
		{"not_a_num", "not_a_num"},
	}

	for _, tt := range tests {
		got := NormalizeVolumeInput(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeVolumeInput(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseVolumeOutputErrors(t *testing.T) {
	invalidOutputs := []string{
		"InvalidOutput",
		"Volume: ",
		"Volume: not_a_float",
	}

	for _, str := range invalidOutputs {
		if _, err := ParseVolumeOutput(str); err == nil {
			t.Errorf("expected error for invalid output %q, got nil", str)
		}
	}
}

func TestVolumeInfoString(t *testing.T) {
	v1 := &VolumeInfo{Percentage: 75, Muted: false}
	if v1.String() != "Volume: 75% [unmuted]" {
		t.Errorf("unexpected string: %s", v1.String())
	}

	v2 := &VolumeInfo{Percentage: 30, Muted: true}
	if v2.String() != "Volume: 30% [MUTED]" {
		t.Errorf("unexpected string: %s", v2.String())
	}
}
