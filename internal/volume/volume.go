package volume

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// VolumeInfo describes the current volume level and mute state.
type VolumeInfo struct {
	Raw        float64 `json:"raw"`
	Percentage int     `json:"percentage"`
	Muted      bool    `json:"muted"`
}

func (v *VolumeInfo) String() string {
	muteStr := "[unmuted]"
	if v.Muted {
		muteStr = "[MUTED]"
	}
	return fmt.Sprintf("Volume: %d%% %s", v.Percentage, muteStr)
}

// ParseVolumeOutput parses the standard text output from wpctl get-volume.
// Example inputs: "Volume: 0.85", "Volume: 0.85 [MUTED]"
func ParseVolumeOutput(output string) (*VolumeInfo, error) {
	trimmed := strings.TrimSpace(output)
	if !strings.HasPrefix(trimmed, "Volume:") {
		return nil, fmt.Errorf("unexpected volume output format: %s", trimmed)
	}

	parts := strings.Fields(trimmed)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed volume output: %s", trimmed)
	}

	valStr := parts[1]
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume value %q: %w", valStr, err)
	}

	info := &VolumeInfo{
		Raw:        val,
		Percentage: int(val*100 + 0.5),
		Muted:      strings.Contains(trimmed, "[MUTED]"),
	}
	return info, nil
}

// NormalizeVolumeInput adjusts user inputs like "+10%" or "-5%" to wpctl compatible format ("10%+", "5%-").
func NormalizeVolumeInput(input string) string {
	in := strings.TrimSpace(input)
	if strings.HasPrefix(in, "+") {
		num := strings.TrimPrefix(in, "+")
		if !strings.HasSuffix(num, "%") {
			num += "%"
		}
		return num + "+"
	}
	if strings.HasPrefix(in, "-") {
		num := strings.TrimPrefix(in, "-")
		if !strings.HasSuffix(num, "%") {
			num += "%"
		}
		return num + "-"
	}
	if !strings.HasSuffix(in, "%") && !strings.Contains(in, ".") {
		// Integer number like 80 -> 80%
		if _, err := strconv.Atoi(in); err == nil {
			return in + "%"
		}
	}
	return in
}

// GetVolume queries current volume and mute state for a given sink.
func GetVolume(sinkID int, sinkName string) (*VolumeInfo, error) {
	if wpctlPath, err := exec.LookPath("wpctl"); err == nil && sinkID > 0 {
		cmd := exec.Command(wpctlPath, "get-volume", strconv.Itoa(sinkID))
		out, err := cmd.Output()
		if err == nil {
			return ParseVolumeOutput(string(out))
		}
	}

	if pactlPath, err := exec.LookPath("pactl"); err == nil && sinkName != "" {
		cmd := exec.Command(pactlPath, "get-sink-volume", sinkName)
		out, err := cmd.Output()
		if err == nil {
			// e.g. "Volume: front-left: 55705 /  85% / -4.24 dB..."
			outStr := string(out)
			if idx := strings.Index(outStr, "/"); idx != -1 {
				after := outStr[idx+1:]
				if pIdx := strings.Index(after, "%"); pIdx != -1 {
					pctStr := strings.TrimSpace(after[:pIdx])
					if pct, err := strconv.Atoi(pctStr); err == nil {
						return &VolumeInfo{
							Raw:        float64(pct) / 100.0,
							Percentage: pct,
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("unable to retrieve volume for sink ID %d (%s)", sinkID, sinkName)
}

// SetVolume sets the volume level for a given sink.
func SetVolume(sinkID int, sinkName string, target string) error {
	normalized := NormalizeVolumeInput(target)

	// 1. Try wpctl set-volume <id> <vol>
	if wpctlPath, err := exec.LookPath("wpctl"); err == nil && sinkID > 0 {
		cmd := exec.Command(wpctlPath, "set-volume", strconv.Itoa(sinkID), normalized)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// 2. Try pactl set-sink-volume <name> <vol>
	if pactlPath, err := exec.LookPath("pactl"); err == nil && sinkName != "" {
		cmd := exec.Command(pactlPath, "set-sink-volume", sinkName, target)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			return fmt.Errorf("pactl set-sink-volume failed: %s (%w)", string(out), err)
		}
	}

	return fmt.Errorf("could not set volume (neither wpctl nor pactl succeeded)")
}

// SetMute sets mute state (true for mute, false for unmute).
func SetMute(sinkID int, sinkName string, mute bool) error {
	muteVal := "0"
	if mute {
		muteVal = "1"
	}

	if wpctlPath, err := exec.LookPath("wpctl"); err == nil && sinkID > 0 {
		cmd := exec.Command(wpctlPath, "set-mute", strconv.Itoa(sinkID), muteVal)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	if pactlPath, err := exec.LookPath("pactl"); err == nil && sinkName != "" {
		cmd := exec.Command(pactlPath, "set-sink-mute", sinkName, muteVal)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			return fmt.Errorf("pactl set-sink-mute failed: %s (%w)", string(out), err)
		}
	}

	return fmt.Errorf("could not set mute state")
}
