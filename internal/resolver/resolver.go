package resolver

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/arthurgray2k/goBlue/internal/bluez"
	"github.com/godbus/dbus/v5"
)

var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

// AmbiguousDeviceError is returned when an identifier matches multiple Bluetooth devices.
type AmbiguousDeviceError struct {
	Query   string
	Matches []*bluez.Device
}

func (e *AmbiguousDeviceError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("multiple devices found matching \"%s\":\n", e.Query))
	for i, m := range e.Matches {
		sb.WriteString(fmt.Sprintf("  [%d] %s (%s)\n", i+1, m.DisplayName(), m.Address))
	}
	sb.WriteString("Please specify the device using its index number or full Bluetooth address.")
	return sb.String()
}

// DeviceNotFoundError is returned when no device matches the query.
type DeviceNotFoundError struct {
	Query string
}

func (e *DeviceNotFoundError) Error() string {
	return fmt.Sprintf("device \"%s\" not found", e.Query)
}

// IsMACAddress returns true if the input looks like a full Bluetooth MAC address.
func IsMACAddress(input string) bool {
	return macRegex.MatchString(strings.TrimSpace(input))
}

// SortDevices deterministically sorts a list of devices (Connected > Paired > Name > Address).
func SortDevices(devices []*bluez.Device) {
	sort.SliceStable(devices, func(i, j int) bool {
		if devices[i].Connected != devices[j].Connected {
			return devices[i].Connected
		}
		if devices[i].Paired != devices[j].Paired {
			return devices[i].Paired
		}
		nameI := strings.ToLower(devices[i].DisplayName())
		nameJ := strings.ToLower(devices[j].DisplayName())
		if nameI != nameJ {
			return nameI < nameJ
		}
		return devices[i].Address < devices[j].Address
	})
}

// ResolveDevice finds a device by index number, MAC address, exact name/alias, short name substring, or MAC suffix.
func ResolveDevice(client bluez.Client, adapterPath dbus.ObjectPath, identifier string) (*bluez.Device, error) {
	cleanID := strings.TrimSpace(identifier)
	if cleanID == "" {
		return nil, fmt.Errorf("device identifier cannot be empty")
	}

	devices, err := client.GetDevices(adapterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	SortDevices(devices)

	// 1. Index number matching (e.g. 1, 2, 3...)
	if idx, err := strconv.Atoi(cleanID); err == nil && !strings.Contains(cleanID, ":") && !strings.Contains(cleanID, "-") {
		if idx >= 1 && idx <= len(devices) {
			return devices[idx-1], nil
		}
		if len(devices) == 0 {
			return nil, fmt.Errorf("no known devices available (index %d invalid)", idx)
		}
		return nil, fmt.Errorf("device index %d out of range (available: 1..%d)", idx, len(devices))
	}

	// 2. Full MAC address matching
	if IsMACAddress(cleanID) {
		normMAC := strings.ToUpper(strings.ReplaceAll(cleanID, "-", ":"))
		for _, d := range devices {
			if strings.ToUpper(d.Address) == normMAC {
				return d, nil
			}
		}
		// If device is not in cache, construct path directly and try fetching
		devPath := bluez.DevicePathFromAddress(adapterPath, normMAC)
		if dev, err := client.GetDevice(devPath); err == nil && dev != nil {
			return dev, nil
		}
		return nil, &DeviceNotFoundError{Query: identifier}
	}

	// 3. Exact Name or Alias matching (case-insensitive)
	var exactMatches []*bluez.Device
	for _, d := range devices {
		if strings.EqualFold(d.Name, cleanID) || strings.EqualFold(d.Alias, cleanID) {
			exactMatches = append(exactMatches, d)
		}
	}
	if len(exactMatches) == 1 {
		return exactMatches[0], nil
	}
	if len(exactMatches) > 1 {
		return nil, &AmbiguousDeviceError{
			Query:   identifier,
			Matches: exactMatches,
		}
	}

	// 4. Substring / Short-name matching (case-insensitive)
	lowerID := strings.ToLower(cleanID)
	var subMatches []*bluez.Device
	for _, d := range devices {
		dispLower := strings.ToLower(d.DisplayName())
		nameLower := strings.ToLower(d.Name)
		aliasLower := strings.ToLower(d.Alias)
		if strings.Contains(dispLower, lowerID) || strings.Contains(nameLower, lowerID) || strings.Contains(aliasLower, lowerID) {
			subMatches = append(subMatches, d)
		}
	}
	if len(subMatches) == 1 {
		return subMatches[0], nil
	}
	if len(subMatches) > 1 {
		return nil, &AmbiguousDeviceError{
			Query:   identifier,
			Matches: subMatches,
		}
	}

	// 5. Short MAC suffix matching (e.g. "e8:d6" or "e8d6")
	cleanHex := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(cleanID, ":", ""), "-", ""))
	if len(cleanHex) >= 4 {
		var suffixMatches []*bluez.Device
		for _, d := range devices {
			addrHex := strings.ToUpper(strings.ReplaceAll(d.Address, ":", ""))
			if strings.HasSuffix(addrHex, cleanHex) {
				suffixMatches = append(suffixMatches, d)
			}
		}
		if len(suffixMatches) == 1 {
			return suffixMatches[0], nil
		}
		if len(suffixMatches) > 1 {
			return nil, &AmbiguousDeviceError{
				Query:   identifier,
				Matches: suffixMatches,
			}
		}
	}

	return nil, &DeviceNotFoundError{Query: identifier}
}
