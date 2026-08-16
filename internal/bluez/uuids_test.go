package bluez

import (
	"testing"
)

func TestNormalizeUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "16-bit hex short format",
			input:    "110b",
			expected: "0000110b-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "16-bit hex with 0x prefix uppercase",
			input:    "0x110A",
			expected: "0000110a-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "32-bit hex format",
			input:    "0000110c",
			expected: "0000110c-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "Full 128-bit lowercase UUID",
			input:    "0000110e-0000-1000-8000-00805f9b34fb",
			expected: "0000110e-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "Full 128-bit uppercase UUID with spaces",
			input:    " 0000111E-0000-1000-8000-00805F9B34FB ",
			expected: "0000111e-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "Custom vendor UUID",
			input:    "df21fe2c-2515-4fdb-8886-f12c4d67927c",
			expected: "df21fe2c-2515-4fdb-8886-f12c4d67927c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeUUID(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeUUID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestServiceNameFromUUID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0000110b-0000-1000-8000-00805f9b34fb", "Audio Sink (A2DP)"},
		{"0000110a-0000-1000-8000-00805f9b34fb", "Audio Source (A2DP)"},
		{"110b", "Audio Sink (A2DP)"},
		{"00001124-0000-1000-8000-00805f9b34fb", "Human Interface Device"},
		{"custom-unknown-uuid", "custom-unknown-uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ServiceNameFromUUID(tt.input)
			if got != tt.expected {
				t.Errorf("ServiceNameFromUUID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
