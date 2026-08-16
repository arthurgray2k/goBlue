package player

import (
	"bytes"
	"os"
	"testing"
)

func TestGenerateTestChimeWAV(t *testing.T) {
	data := GenerateTestChimeWAV()
	if len(data) < 44 {
		t.Fatalf("WAV file too short: %d bytes", len(data))
	}

	// Verify RIFF header
	if !bytes.Equal(data[:4], []byte("RIFF")) {
		t.Errorf("expected RIFF header, got %s", string(data[:4]))
	}
	if !bytes.Equal(data[8:12], []byte("WAVE")) {
		t.Errorf("expected WAVE header, got %s", string(data[8:12]))
	}
	if !bytes.Equal(data[12:16], []byte("fmt ")) {
		t.Errorf("expected fmt subchunk, got %s", string(data[12:16]))
	}
}

func TestPlayFileNonExistent(t *testing.T) {
	err := PlayFile(1, "test_sink", "/path/to/non_existent_audio_file.wav")
	if err == nil {
		t.Fatalf("expected error for non-existent file")
	}
}

func TestPlayFileTemporary(t *testing.T) {
	tmp, err := os.CreateTemp("", "test-*.wav")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	wavData := GenerateTestChimeWAV()
	_, _ = tmp.Write(wavData)
	_ = tmp.Close()

	// Verify stat
	if _, err := os.Stat(tmp.Name()); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}
