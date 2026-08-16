package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arthurgray2k/goBlue/internal/audio"
	"github.com/arthurgray2k/goBlue/internal/bluez"
	"github.com/arthurgray2k/goBlue/internal/connector"
	"github.com/arthurgray2k/goBlue/internal/player"
	"github.com/arthurgray2k/goBlue/internal/render"
	"github.com/arthurgray2k/goBlue/internal/resolver"
	"github.com/arthurgray2k/goBlue/internal/volume"
)

const version = "1.0.0"

func printUsage() {
	fmt.Println(`goBlue — Linux Bluetooth & Audio CLI

Usage:
  goBlue <command> [arguments] [flags]

Discovery & Status Commands:
  adapter                       Show local Bluetooth adapter details
  adapter set-name <name>       Set broadcast/projected Bluetooth name
  adapter set-name --reset      Reset broadcast name to system default
  devices                       List all known Bluetooth devices
  scan                          Scan for nearby Bluetooth devices
  status                        Show adapter status and connected devices
  info <device>                 Show detailed information and audio capabilities

Device Management Commands:
  connect <device>              Connect to a Bluetooth device (manual session)
  disconnect <device>           Disconnect from a Bluetooth device
  pair <device>                 Pair with a Bluetooth device
  trust <device>                Trust a Bluetooth device (auto-reconnect)
  untrust <device>              Untrust a Bluetooth device

Volume & Playback Commands:
  volume <device> [level]       Inspect or set volume (e.g. 80%, +10%, -5%)
  mute <device>                 Mute audio output on device
  unmute <device>               Unmute audio output on device
  play-test <device>            Play acoustic test chime to verify sound output
  play <device> <file>          Stream audio file (WAV/FLAC/MP3) directly to device

Audio Routing Commands:
  audio status                  Show Bluetooth audio sinks and default status
  audio info <device>           Inspect Bluetooth audio roles (A2DP Sink/Source)
  audio set-default <device>    Connect and route system audio to device
  audio volume <device> [level] Alias for volume inspection/adjustment
  audio mute <device>           Alias for mute
  audio unmute <device>         Alias for unmute
  audio play-test <device>      Alias for play-test
  audio play <device> <file>    Alias for play

Global Flags:
  -a, --adapter <name>          Specify Bluetooth adapter (default: auto)
      --json                    Output structured JSON
      --timeout <duration>      Scan duration (default: 8s)
  -h, --help                    Show this help message
  -v, --version                 Show version`)
}

func printAudioUsage() {
	fmt.Println(`goBlue Audio & Volume Management

Usage:
  goBlue audio <subcommand> [arguments]
  goBlue volume <device> [level]
  goBlue mute <device>
  goBlue unmute <device>
  goBlue play-test <device>
  goBlue play <device> <file>

Subcommands:
  status                        Show Bluetooth audio sinks and default status
  info <device>                 Inspect Bluetooth audio roles (A2DP Sink/Source)
  set-default <device>          Connect and set device as default audio output
  volume <device> [level]       Inspect or adjust volume level:
                                  goBlue volume "Sony WH-1000XM5"       (get volume)
                                  goBlue volume "Sony WH-1000XM5" 80%   (set to 80%)
                                  goBlue volume "Sony WH-1000XM5" +10%  (increase by 10%)
                                  goBlue volume "Sony WH-1000XM5" -5%   (decrease by 5%)
  mute <device>                 Mute audio output on device
  unmute <device>               Unmute audio output on device
  play-test <device>            Play test acoustic chime through device
  play <device> <file>          Stream audio file (WAV/FLAC/MP3) directly to device`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Top-level arguments
	firstArg := os.Args[1]
	if firstArg == "-h" || firstArg == "--help" || firstArg == "help" {
		printUsage()
		return
	}
	if firstArg == "-v" || firstArg == "--version" || firstArg == "version" {
		fmt.Printf("goBlue version %s\n", version)
		return
	}

	// Parse flags that can appear anywhere or before/after subcommands
	var adapterName string
	var jsonOutput bool
	var timeoutStr string

	// Extract subcommand and remaining args
	var subcmd string
	var cmdArgs []string

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--json":
			jsonOutput = true
		case arg == "-a" || arg == "--adapter":
			if i+1 < len(os.Args) {
				i++
				adapterName = os.Args[i]
			}
		case strings.HasPrefix(arg, "--adapter="):
			adapterName = strings.TrimPrefix(arg, "--adapter=")
		case arg == "--timeout":
			if i+1 < len(os.Args) {
				i++
				timeoutStr = os.Args[i]
			}
		case strings.HasPrefix(arg, "--timeout="):
			timeoutStr = strings.TrimPrefix(arg, "--timeout=")
		case subcmd == "":
			subcmd = arg
		default:
			cmdArgs = append(cmdArgs, arg)
		}
	}

	timeout := 8 * time.Second
	if timeoutStr != "" {
		if d, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = d
		}
	}

	client, err := bluez.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to BlueZ: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	audioClient := audio.NewClient()
	conn := connector.NewConnector(client)

	switch subcmd {
	case "help", "--help", "-h":
		if len(cmdArgs) > 0 {
			switch cmdArgs[0] {
			case "audio", "volume", "sound", "mute", "play":
				printAudioUsage()
				return
			}
		}
		printUsage()
		return
	case "adapter":
		handleAdapter(client, adapterName, cmdArgs, jsonOutput)
	case "devices":
		handleDevices(client, adapterName, jsonOutput)
	case "scan":
		handleScan(client, adapterName, timeout, jsonOutput)
	case "status":
		handleStatus(client, audioClient, adapterName, jsonOutput)
	case "info":
		if len(cmdArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: missing device argument.\nUsage: goBlue info <device>")
			os.Exit(1)
		}
		handleInfo(client, adapterName, cmdArgs[0], jsonOutput)
	case "pair":
		if len(cmdArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: missing device argument.\nUsage: goBlue pair <device>")
			os.Exit(1)
		}
		handlePair(client, adapterName, cmdArgs[0])
	case "trust":
		if len(cmdArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: missing device argument.\nUsage: goBlue trust <device>")
			os.Exit(1)
		}
		handleTrust(client, adapterName, cmdArgs[0], true)
	case "untrust":
		if len(cmdArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: missing device argument.\nUsage: goBlue untrust <device>")
			os.Exit(1)
		}
		handleTrust(client, adapterName, cmdArgs[0], false)
	case "connect":
		if len(cmdArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: missing device argument.\nUsage: goBlue connect <device>")
			os.Exit(1)
		}
		handleConnect(client, conn, adapterName, cmdArgs[0])
	case "disconnect":
		if len(cmdArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: missing device argument.\nUsage: goBlue disconnect <device>")
			os.Exit(1)
		}
		handleDisconnect(client, conn, adapterName, cmdArgs[0])
	case "volume":
		handleAudio(client, audioClient, conn, adapterName, append([]string{"volume"}, cmdArgs...), jsonOutput)
	case "mute":
		handleAudio(client, audioClient, conn, adapterName, append([]string{"mute"}, cmdArgs...), jsonOutput)
	case "unmute":
		handleAudio(client, audioClient, conn, adapterName, append([]string{"unmute"}, cmdArgs...), jsonOutput)
	case "play-test", "test":
		handleAudio(client, audioClient, conn, adapterName, append([]string{"play-test"}, cmdArgs...), jsonOutput)
	case "play":
		handleAudio(client, audioClient, conn, adapterName, append([]string{"play"}, cmdArgs...), jsonOutput)
	case "audio":
		handleAudio(client, audioClient, conn, adapterName, cmdArgs, jsonOutput)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Run 'goBlue --help' for available commands.\n", subcmd)
		os.Exit(1)
	}
}

func getAdapter(client bluez.Client, adapterName string) (*bluez.Adapter, error) {
	adapter, err := client.GetDefaultAdapter(adapterName)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return adapter, nil
}

func handleAdapter(client bluez.Client, adapterName string, args []string, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(args) > 0 {
		sub := args[0]
		switch sub {
		case "set-name", "name":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: missing name argument.\nUsage: goBlue adapter set-name <name> (or --reset)")
				os.Exit(1)
			}
			newName := args[1]
			if newName == "--reset" || newName == "-r" {
				if err := client.SetAdapterAlias(adapter.Path, ""); err != nil {
					fmt.Fprintf(os.Stderr, "✗ Failed to reset adapter name: %v\n", err)
					os.Exit(1)
				}
				time.Sleep(100 * time.Millisecond)
				fmt.Printf("✓ Adapter name reset to system default (\"%s\")\n", adapter.Name)
				return
			}
			if err := client.SetAdapterAlias(adapter.Path, newName); err != nil {
				fmt.Fprintf(os.Stderr, "✗ Failed to set adapter name: %v\n", err)
				os.Exit(1)
			}
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("✓ Adapter name set to \"%s\"\n", newName)
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown adapter subcommand '%s'.\nUsage: goBlue adapter [set-name <name>]\n", sub)
			os.Exit(1)
		}
	}

	if jsonOutput {
		_ = render.RenderJSON(os.Stdout, adapter)
		return
	}
	render.RenderAdapter(os.Stdout, adapter)
}

func handleDevices(client bluez.Client, adapterName string, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	devices, err := client.GetDevices(adapter.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching devices: %v\n", err)
		os.Exit(1)
	}
	resolver.SortDevices(devices)
	if jsonOutput {
		_ = render.RenderJSON(os.Stdout, devices)
		return
	}
	render.RenderDevices(os.Stdout, devices)
}

func handleScan(client bluez.Client, adapterName string, timeout time.Duration, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if !jsonOutput {
		fmt.Printf("Scanning for Bluetooth devices on %s (%v)...\n\n", adapter.ID, timeout)
	}

	devices, err := client.Scan(adapter.Path, timeout, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		_ = render.RenderJSON(os.Stdout, devices)
		return
	}
	render.RenderScanResults(os.Stdout, devices)
}

func handleStatus(client bluez.Client, audioClient *audio.Client, adapterName string, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	devices, err := client.GetDevices(adapter.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching devices: %v\n", err)
		os.Exit(1)
	}

	var connectedDevs []*bluez.Device
	for _, d := range devices {
		if d.Connected {
			connectedDevs = append(connectedDevs, d)
		}
	}

	var sinks []*audio.AudioSink
	if audioClient.IsAvailable() {
		sinks, _ = audioClient.GetBluetoothSinks()
	}

	sinkByAddr := make(map[string]*audio.AudioSink)
	for _, s := range sinks {
		if s.BluetoothAddress != "" {
			sinkByAddr[s.BluetoothAddress] = s
		}
	}

	extras := make([]*render.DeviceStatusExtra, 0)
	for _, d := range connectedDevs {
		normAddr := strings.ToUpper(d.Address)
		extra := &render.DeviceStatusExtra{
			Device:      d,
			HasAudioCap: d.AudioCaps.CanReceiveAudio(),
		}
		if sink, ok := sinkByAddr[normAddr]; ok {
			extra.AudioSink = sink
			extra.Profile = sink.Profile
			extra.IsDefault = sink.IsDefault
		}
		extras = append(extras, extra)
	}

	if jsonOutput {
		statusData := map[string]interface{}{
			"adapter":           adapter,
			"connected_devices": extras,
		}
		_ = render.RenderJSON(os.Stdout, statusData)
		return
	}

	render.RenderStatus(os.Stdout, adapter, extras)
}

func handleInfo(client bluez.Client, adapterName, identifier string, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dev, err := resolver.ResolveDevice(client, adapter.Path, identifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if jsonOutput {
		_ = render.RenderJSON(os.Stdout, dev)
		return
	}
	render.RenderDeviceInfo(os.Stdout, dev)
}

func handlePair(client bluez.Client, adapterName, identifier string) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dev, err := resolver.ResolveDevice(client, adapter.Path, identifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Device found")
	if dev.Paired {
		fmt.Println("Already paired")
		return
	}

	fmt.Println("Pairing...")
	if err := client.Pair(dev.Path); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Pairing failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Paired successfully")
}

func handleTrust(client bluez.Client, adapterName, identifier string, trusted bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dev, err := resolver.ResolveDevice(client, adapter.Path, identifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := client.SetTrusted(dev.Path, trusted); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to update trust: %v\n", err)
		os.Exit(1)
	}

	if trusted {
		fmt.Printf("✓ Trusted %s\n", dev.Address)
	} else {
		fmt.Printf("✓ Untrusted %s\n", dev.Address)
	}
}

func handleConnect(client bluez.Client, conn *connector.Connector, adapterName, identifier string) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Device found.")

	res, err := conn.Connect(adapter.Path, identifier)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Connection failed: %v\n", err)
		os.Exit(1)
	}

	if res.HasA2DPSink {
		fmt.Println("A2DP Sink detected.")
	} else if len(res.Device.UUIDs) > 0 {
		fmt.Println()
		fmt.Println("A2DP Sink not detected.")
		fmt.Println("The device may support Bluetooth, but it does not appear to accept")
		fmt.Println("audio from this Linux machine.")
		fmt.Println()
	}

	if res.AlreadyConnected {
		fmt.Println("Already connected")
		return
	}

	fmt.Println("✓ Connected")
}

func handleDisconnect(client bluez.Client, conn *connector.Connector, adapterName, identifier string) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	res, err := conn.Disconnect(adapter.Path, identifier)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Disconnect failed: %v\n", err)
		os.Exit(1)
	}

	if res.AlreadyDisconnected {
		fmt.Println("✓ Already disconnected")
		return
	}

	fmt.Println("✓ Disconnected")
}

func handleAudio(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName string, args []string, jsonOutput bool) {
	if len(args) == 0 || args[0] == "status" {
		handleAudioStatus(client, audioClient, adapterName, jsonOutput)
		return
	}

	sub := args[0]
	switch sub {
	case "help", "--help", "-h":
		printAudioUsage()
		return
	case "info":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			fmt.Fprintln(os.Stderr, "Usage: goBlue audio info <device>")
			os.Exit(1)
		}
		handleAudioInfo(client, audioClient, adapterName, args[1], jsonOutput)
	case "set-default":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			fmt.Fprintln(os.Stderr, "Usage: goBlue audio set-default <device>")
			os.Exit(1)
		}
		handleAudioSetDefault(client, audioClient, conn, adapterName, args[1])
	case "volume":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			fmt.Println(`Usage: goBlue volume <device> [level]

Examples:
  goBlue volume "Sony WH-1000XM5"       # Get current volume level
  goBlue volume "Sony WH-1000XM5" 80%   # Set volume to 80%
  goBlue volume "Sony WH-1000XM5" +10%  # Increase volume by 10%
  goBlue volume "Sony WH-1000XM5" -5%   # Decrease volume by 5%`)
			return
		}
		var targetVol string
		if len(args) >= 3 {
			targetVol = args[2]
		}
		handleAudioVolume(client, audioClient, conn, adapterName, args[1], targetVol, jsonOutput)
	case "mute":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			fmt.Fprintln(os.Stderr, "Usage: goBlue mute <device>")
			os.Exit(1)
		}
		handleAudioMute(client, audioClient, conn, adapterName, args[1], true)
	case "unmute":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			fmt.Fprintln(os.Stderr, "Usage: goBlue unmute <device>")
			os.Exit(1)
		}
		handleAudioMute(client, audioClient, conn, adapterName, args[1], false)
	case "play-test", "test":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			fmt.Fprintln(os.Stderr, "Usage: goBlue play-test <device>")
			os.Exit(1)
		}
		handleAudioPlayTest(client, audioClient, conn, adapterName, args[1])
	case "play":
		if len(args) < 3 || args[1] == "--help" || args[1] == "-h" {
			fmt.Fprintln(os.Stderr, "Usage: goBlue play <device> <audio-file>")
			os.Exit(1)
		}
		handleAudioPlayFile(client, audioClient, conn, adapterName, args[1], args[2])
	default:
		fmt.Fprintf(os.Stderr, "Unknown audio subcommand '%s'. Run 'goBlue audio --help' for usage.\n", sub)
		os.Exit(1)
	}
}

func getConnectedSink(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName, identifier string) (*bluez.Device, *audio.AudioSink) {
	if !audioClient.IsAvailable() {
		fmt.Fprintln(os.Stderr, "Error: PipeWire / WirePlumber audio management tools not available on this system.")
		os.Exit(1)
	}

	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dev, err := conn.EnsureConnected(adapter.Path, identifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sink, err := audioClient.FindSinkByBluetoothAddress(dev.Address)
	if err != nil {
		for i := 0; i < 4; i++ {
			time.Sleep(500 * time.Millisecond)
			sink, err = audioClient.FindSinkByBluetoothAddress(dev.Address)
			if err == nil && sink != nil {
				break
			}
		}
	}

	if err != nil || sink == nil {
		fmt.Fprintf(os.Stderr, "✗ No PipeWire audio sink found for %s (%s).\n", dev.DisplayName(), dev.Address)
		os.Exit(1)
	}

	return dev, sink
}

func handleAudioVolume(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName, identifier, targetVol string, jsonOutput bool) {
	dev, sink := getConnectedSink(client, audioClient, conn, adapterName, identifier)

	if targetVol != "" {
		if err := volume.SetVolume(sink.ID, sink.Name, targetVol); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to set volume: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Volume set to %s for %s\n", targetVol, dev.DisplayName())
		return
	}

	volInfo, err := volume.GetVolume(sink.ID, sink.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to get volume: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		_ = render.RenderJSON(os.Stdout, volInfo)
		return
	}
	fmt.Printf("%s\n", volInfo.String())
}

func handleAudioMute(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName, identifier string, mute bool) {
	dev, sink := getConnectedSink(client, audioClient, conn, adapterName, identifier)

	if err := volume.SetMute(sink.ID, sink.Name, mute); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to update mute state: %v\n", err)
		os.Exit(1)
	}

	if mute {
		fmt.Printf("✓ Muted %s\n", dev.DisplayName())
	} else {
		fmt.Printf("✓ Unmuted %s\n", dev.DisplayName())
	}
}

func handleAudioPlayTest(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName, identifier string) {
	dev, sink := getConnectedSink(client, audioClient, conn, adapterName, identifier)

	fmt.Printf("Playing test chime to %s...\n", dev.DisplayName())
	if err := player.PlayTone(sink.ID, sink.Name); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Audio playback failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Playback complete")
}

func handleAudioPlayFile(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName, identifier, filePath string) {
	dev, sink := getConnectedSink(client, audioClient, conn, adapterName, identifier)

	fmt.Printf("Playing %s to %s...\n", filePath, dev.DisplayName())
	if err := player.PlayFile(sink.ID, sink.Name, filePath); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Audio playback failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Playback complete")
}

func handleAudioSetDefault(client bluez.Client, audioClient *audio.Client, conn *connector.Connector, adapterName, identifier string) {
	dev, sink := getConnectedSink(client, audioClient, conn, adapterName, identifier)

	if err := audioClient.SetDefaultSink(sink); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to set default audio sink: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Device %s connected\n", dev.DisplayName())
	fmt.Println("✓ Bluetooth audio sink available")
	fmt.Println("✓ Default audio output changed")
}

func handleAudioStatus(client bluez.Client, audioClient *audio.Client, adapterName string, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	devices, err := client.GetDevices(adapter.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching devices: %v\n", err)
		os.Exit(1)
	}

	items, err := audioClient.GetAudioStatus(devices)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Audio status error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		_ = render.RenderJSON(os.Stdout, items)
		return
	}
	render.RenderAudioStatus(os.Stdout, items)
}

func handleAudioInfo(client bluez.Client, audioClient *audio.Client, adapterName, identifier string, jsonOutput bool) {
	adapter, err := getAdapter(client, adapterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dev, err := resolver.ResolveDevice(client, adapter.Path, identifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var hasPipeWireSink bool
	if audioClient.IsAvailable() {
		if sink, _ := audioClient.FindSinkByBluetoothAddress(dev.Address); sink != nil {
			hasPipeWireSink = true
		}
	}

	if jsonOutput {
		infoData := map[string]interface{}{
			"device":             dev,
			"has_pipewire_sink":  hasPipeWireSink,
			"audio_capabilities": dev.AudioCaps,
			"can_receive_audio":  dev.AudioCaps.CanReceiveAudio(),
		}
		_ = render.RenderJSON(os.Stdout, infoData)
		return
	}

	render.RenderAudioInfo(os.Stdout, dev, hasPipeWireSink)
}
