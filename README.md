# goBlue

A fast, lightweight Linux CLI in Go to discover, inspect, pair, trust, connect, disconnect, manage volume, and stream audio to Bluetooth audio devices (headphones, earbuds, speakers, soundbars, TVs) using **BlueZ D-Bus** and **PipeWire / WirePlumber**.

---

## Architecture

```text
goBlue
   |
   +--> D-Bus (System Bus) --> BlueZ (org.bluez)
   |                             |
   |                             v
   |                      Linux Bluetooth Subsystem / Adapter
   |
   +--> PipeWire / WirePlumber (Audio Routing & Playback)
                                 |
                                 v
                          Bluetooth Audio Sink
```

`goBlue` operates at the user-space controller level:
- Communicates directly with BlueZ over the native Linux System D-Bus.
- Coordinates with PipeWire and WirePlumber for discovering audio sinks, inspecting audio profiles (A2DP / HFP / HSP), managing volume, streaming audio, and switching the system default audio sink.
- Zero CGo dependencies: Implemented in pure Go using standard Linux system interfaces.
- Local-only: No telemetry, no background daemons, no cloud services.

---

## Features & Commands

### 1. Volume & Audio Playback
| Command | Description | Example |
|---|---|---|
| `volume <device>` | Check current volume level and mute state | `goBlue volume "Sony WH-1000XM5"` |
| `volume <device> <level>` | Set specific volume level (0-100%) | `goBlue volume "Sony WH-1000XM5" 80%` |
| `volume <device> +N% / -N%` | Relative volume adjustments | `goBlue volume "Sony WH-1000XM5" +10%` |
| `mute <device>` | Mute output on device | `goBlue mute "Sony WH-1000XM5"` |
| `unmute <device>` | Unmute output on device | `goBlue unmute "Sony WH-1000XM5"` |
| `play-test <device>` | Play acoustic test chime (sound check) | `goBlue play-test "Sony WH-1000XM5"` |
| `play <device> <file>` | Stream audio file (WAV/FLAC/MP3) | `goBlue play "Sony WH-1000XM5" song.wav` |

### 2. Device Connection & Management
| Command | Description | Example |
|---|---|---|
| `connect <device>` | Connect without altering permanent trust or default output | `goBlue connect "Sony WH-1000XM5"` |
| `disconnect <device>` | Disconnect active connection | `goBlue disconnect "Sony WH-1000XM5"` |
| `pair <device>` | Pair with device (with PIN/passkey agent) | `goBlue pair "Sony WH-1000XM5"` |
| `trust <device>` | Trust device for automatic reconnection | `goBlue trust "Sony WH-1000XM5"` |
| `untrust <device>` | Remove automatic reconnection trust | `goBlue untrust "Sony WH-1000XM5"` |

### 3. Discovery & Status
| Command | Description | Example |
|---|---|---|
| `adapter` | Show local Bluetooth adapter details | `goBlue adapter` |
| `adapter set-name <name>` | Change broadcast Bluetooth name | `goBlue adapter set-name "My Laptop"` |
| `adapter set-name --reset` | Reset broadcast name to hostname | `goBlue adapter set-name --reset` |
| `devices` | List all known / paired devices | `goBlue devices` |
| `scan` | Discover nearby Bluetooth devices | `goBlue scan --timeout 8s` |
| `status` | Overview of adapter and connected devices | `goBlue status` |
| `info <device>` | Show device details and audio roles | `goBlue info "Sony WH-1000XM5"` |

### 4. Audio Routing & Inspection
| Command | Description | Example |
|---|---|---|
| `audio status` | List audio sinks, active profiles & default marker | `goBlue audio status` |
| `audio info <device>` | Check A2DP Sink capability to receive PC audio | `goBlue audio info "Living Room TV"` |
| `audio set-default <device>` | Connect and route all desktop audio to device | `goBlue audio set-default "Sony WH-1000XM5"` |

---

## Installation & Build

Requires Go 1.26+ installed on Linux:

```bash
git clone git@github.com:arthurgray2k/goBlue.git
cd goBlue
make
```

Or manually:

```bash
go build -o goBlue ./cmd/goBlue
```

---

## Quick Start Workflows

### Workflow 1: Quick Connect & Adjust Volume / Play Music
*(Connect and listen without altering permanent trust or default sound routing)*
```bash
# 1. Connect
./goBlue connect "Sony WH-1000XM5"

# 2. Check and adjust volume
./goBlue volume "Sony WH-1000XM5"
./goBlue volume "Sony WH-1000XM5" 75%
./goBlue volume "Sony WH-1000XM5" +10%

# 3. Test sound output with a test chime
./goBlue play-test "Sony WH-1000XM5"

# 4. Stream a music file directly to the headphones
./goBlue play "Sony WH-1000XM5" music.wav
```

### Workflow 2: Pair, Trust, and Route All Desktop Audio
*(Route browser, Spotify, and system audio to the Bluetooth device)*
```bash
# 1. Scan for nearby devices
./goBlue scan

# 2. Pair and trust for automatic reconnects
./goBlue pair "Sony WH-1000XM5"
./goBlue trust "Sony WH-1000XM5"

# 3. Connect and set as default audio output
./goBlue audio set-default "Sony WH-1000XM5"

# 4. Check status
./goBlue audio status
```

For complete command references and options, see [USAGE.md](USAGE.md).

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
