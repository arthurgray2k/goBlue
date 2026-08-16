package player

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// PlayTone generates a test audio chime and plays it directly to the target sink.
func PlayTone(sinkID int, sinkName string) error {
	wavData := GenerateTestChimeWAV()

	tmpFile, err := os.CreateTemp("", "goBlue-tone-*.wav")
	if err != nil {
		return fmt.Errorf("failed to create temporary audio file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(wavData); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write tone data: %w", err)
	}
	_ = tmpFile.Close()

	return PlayFile(sinkID, sinkName, tmpFile.Name())
}

// PlayFile streams a given audio file directly to the specified PipeWire/Pulse sink.
func PlayFile(sinkID int, sinkName string, filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("audio file not found: %w", err)
	}

	// 1. Try pw-play --target <sinkID> <filePath>
	if pwPlayPath, err := exec.LookPath("pw-play"); err == nil && sinkID > 0 {
		cmd := exec.Command(pwPlayPath, "--target", strconv.Itoa(sinkID), filePath)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// 2. Try paplay --device=<sinkName> <filePath>
	if paplayPath, err := exec.LookPath("paplay"); err == nil && sinkName != "" {
		cmd := exec.Command(paplayPath, "--device="+sinkName, filePath)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			return fmt.Errorf("paplay playback failed: %s (%w)", string(out), err)
		}
	}

	// 3. Fallback generic pw-play
	if pwPlayPath, err := exec.LookPath("pw-play"); err == nil {
		cmd := exec.Command(pwPlayPath, filePath)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			return fmt.Errorf("pw-play playback failed: %s (%w)", string(out), err)
		}
	}

	return fmt.Errorf("no audio playback tool (pw-play or paplay) available")
}
