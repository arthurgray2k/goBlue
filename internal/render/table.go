package render

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/arthurgray2k/goBlue/internal/audio"
	"github.com/arthurgray2k/goBlue/internal/bluez"
)

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// RenderAdapter outputs adapter details.
func RenderAdapter(w io.Writer, a *bluez.Adapter) {
	fmt.Fprintln(w, "BLUETOOTH ADAPTER")
	fmt.Fprintln(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Adapter:\t%s\n", a.ID)
	if a.Address != "" {
		fmt.Fprintf(tw, "Address:\t%s\n", a.Address)
	}
	if a.Name != "" {
		fmt.Fprintf(tw, "Name:\t%s\n", a.Name)
	}
	if a.Alias != "" && a.Alias != a.Name {
		fmt.Fprintf(tw, "Alias:\t%s\n", a.Alias)
	}
	fmt.Fprintf(tw, "Powered:\t%s\n", boolToYesNo(a.Powered))
	fmt.Fprintf(tw, "Discoverable:\t%s\n", boolToYesNo(a.Discoverable))
	fmt.Fprintf(tw, "Pairable:\t%s\n", boolToYesNo(a.Pairable))
	fmt.Fprintf(tw, "Discovering:\t%s\n", boolToYesNo(a.Discovering))
	tw.Flush()
}

// RenderDevices outputs known devices table with index numbers.
func RenderDevices(w io.Writer, devices []*bluez.Device) {
	fmt.Fprintln(w, "KNOWN BLUETOOTH DEVICES")
	fmt.Fprintln(w)
	if len(devices) == 0 {
		fmt.Fprintln(w, "No known Bluetooth devices.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 4, ' ', 0)
	fmt.Fprintln(tw, "#\tNAME\tADDRESS\tPAIRED\tTRUSTED\tCONNECTED")
	for i, d := range devices {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n",
			i+1,
			d.DisplayName(),
			d.Address,
			boolToYesNo(d.Paired),
			boolToYesNo(d.Trusted),
			boolToYesNo(d.Connected),
		)
	}
	tw.Flush()
}

// RenderScanResults outputs discovered devices table with index numbers.
func RenderScanResults(w io.Writer, devices []*bluez.Device) {
	fmt.Fprintln(w, "BLUETOOTH DEVICES")
	fmt.Fprintln(w)
	if len(devices) == 0 {
		fmt.Fprintln(w, "No Bluetooth devices found.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 4, ' ', 0)
	fmt.Fprintln(tw, "#\tNAME\tADDRESS\tRSSI\tSTATE")
	for i, d := range devices {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			i+1,
			d.DisplayName(),
			d.Address,
			d.RSSIString(),
			d.StateString(),
		)
	}
	tw.Flush()
}

// DeviceStatusExtra holds correlated audio sink information for status view.
type DeviceStatusExtra struct {
	Device      *bluez.Device
	AudioSink   *audio.AudioSink
	HasAudioCap bool
	Profile     string
	IsDefault   bool
}

// RenderStatus outputs system status overview.
func RenderStatus(w io.Writer, a *bluez.Adapter, connected []*DeviceStatusExtra) {
	fmt.Fprintln(w, "BLUETOOTH STATUS")
	fmt.Fprintln(w)
	if a != nil {
		fmt.Fprintf(w, "Adapter:      %s\n", a.ID)
		fmt.Fprintf(w, "Powered:      %s\n", boolToYesNo(a.Powered))
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Connected devices:")
	fmt.Fprintln(w)

	if len(connected) == 0 {
		fmt.Fprintln(w, "No devices currently connected.")
		return
	}

	for i, extra := range connected {
		d := extra.Device
		fmt.Fprintf(w, "[%d] %s\n", i+1, d.DisplayName())
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "    Address:\t%s\n", d.Address)
		fmt.Fprintf(tw, "    Paired:\t%s\n", boolToYesNo(d.Paired))
		fmt.Fprintf(tw, "    Trusted:\t%s\n", boolToYesNo(d.Trusted))
		fmt.Fprintf(tw, "    Connected:\t%s\n", boolToYesNo(d.Connected))
		fmt.Fprintf(tw, "    Audio:\t%s\n", boolToYesNo(extra.HasAudioCap || extra.AudioSink != nil))
		if extra.Profile != "" {
			fmt.Fprintf(tw, "    Profile:\t%s\n", extra.Profile)
		} else if extra.Device.AudioCaps.HasA2DPSink {
			fmt.Fprintf(tw, "    Profile:\tA2DP\n")
		}
		if extra.AudioSink != nil {
			fmt.Fprintf(tw, "    Default:\t%s\n", boolToYesNo(extra.IsDefault))
		}
		tw.Flush()
		if i < len(connected)-1 {
			fmt.Fprintln(w)
		}
	}
}

// RenderDeviceInfo outputs detailed device info.
func RenderDeviceInfo(w io.Writer, d *bluez.Device) {
	fmt.Fprintln(w, "DEVICE")
	fmt.Fprintln(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Name:\t%s\n", d.DisplayName())
	fmt.Fprintf(tw, "Address:\t%s\n", d.Address)
	fmt.Fprintf(tw, "Type:\t%s\n", d.TypeString())
	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "Paired:\t%s\n", boolToYesNo(d.Paired))
	fmt.Fprintf(tw, "Trusted:\t%s\n", boolToYesNo(d.Trusted))
	fmt.Fprintf(tw, "Connected:\t%s\n", boolToYesNo(d.Connected))
	if d.Blocked {
		fmt.Fprintf(tw, "Blocked:\t%s\n", boolToYesNo(d.Blocked))
	}
	tw.Flush()

	fmt.Fprintln(w)
	fmt.Fprintln(w, "AUDIO CAPABILITIES")
	fmt.Fprintln(w)
	twAudio := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twAudio, "A2DP Sink:\t%s\n", boolToYesNo(d.AudioCaps.HasA2DPSink))
	fmt.Fprintf(twAudio, "A2DP Source:\t%s\n", boolToYesNo(d.AudioCaps.HasA2DPSource))
	fmt.Fprintf(twAudio, "HFP/HSP:\t%s\n", boolToYesNo(d.AudioCaps.HasHFP || d.AudioCaps.HasHSP))
	fmt.Fprintf(twAudio, "AVRCP:\t%s\n", boolToYesNo(d.AudioCaps.HasAVRCP))
	twAudio.Flush()
}

// RenderAudioStatus outputs the audio status table with index numbers.
func RenderAudioStatus(w io.Writer, items []*audio.AudioStatusItem) {
	fmt.Fprintln(w, "BLUETOOTH AUDIO")
	fmt.Fprintln(w)
	if len(items) == 0 {
		fmt.Fprintln(w, "No Bluetooth audio devices found.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 4, ' ', 0)
	fmt.Fprintln(tw, "#\tDEVICE\tSTATE\tPROFILE\tDEFAULT")
	for i, item := range items {
		prof := item.Profile
		if prof == "" {
			prof = "-"
		}
		defStr := "-"
		if item.State == "connected" {
			defStr = boolToYesNo(item.IsDefault)
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			i+1,
			item.DeviceName,
			item.State,
			prof,
			defStr,
		)
	}
	tw.Flush()
}

// RenderAudioInfo outputs detailed audio role capability for a single device.
func RenderAudioInfo(w io.Writer, d *bluez.Device, hasPipeWireSink bool) {
	fmt.Fprintln(w, "AUDIO CAPABILITIES")
	fmt.Fprintln(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Device:\t%s\n", d.DisplayName())
	fmt.Fprintf(tw, "Address:\t%s\n", d.Address)
	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "A2DP Sink:\t%s\n", boolToYesNo(d.AudioCaps.HasA2DPSink))
	fmt.Fprintf(tw, "A2DP Source:\t%s\n", boolToYesNo(d.AudioCaps.HasA2DPSource))
	fmt.Fprintf(tw, "HFP/HSP:\t%s\n", boolToYesNo(d.AudioCaps.HasHFP || d.AudioCaps.HasHSP))
	fmt.Fprintf(tw, "AVRCP:\t%s\n", boolToYesNo(d.AudioCaps.HasAVRCP))
	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "Can receive audio from this Linux machine:\t%s\n", boolToYesNo(d.AudioCaps.CanReceiveAudio()))
	if hasPipeWireSink {
		fmt.Fprintf(tw, "PipeWire audio sink available:\tyes\n")
	} else if d.Connected {
		fmt.Fprintf(tw, "PipeWire audio sink available:\tno\n")
	}
	tw.Flush()
}
