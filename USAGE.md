# USAGE: goBlue

`goBlue` provides a fast command-line toolkit for discovering, connecting, controlling volume, and streaming audio to Bluetooth devices on Linux.

---

## Device Shortcuts & Index Numbers

All device commands accept **Index Numbers (`1`, `2`, `3`)**, **Short Names / Substrings (`buds`, `sony`)**, **MAC Suffixes (`e8:d6`)**, or **Full MAC Addresses**:

```bash
# View devices with their index numbers (#):
$ goBlue devices

KNOWN BLUETOOTH DEVICES

#   NAME                     ADDRESS              PAIRED    TRUSTED    CONNECTED
1   realme Buds T200 Lite    28:04:C6:73:E8:D6    yes       yes        yes
2   Sony WH-1000XM5          AA:BB:CC:DD:EE:FF    yes       yes        no

# Use device numbers directly:
$ goBlue connect 1               # Connect device #1
$ goBlue volume 1 80%            # Set volume on device #1 to 80%
$ goBlue play-test 1             # Play test sound on device #1
$ goBlue audio set-default 1     # Set device #1 as default audio output
$ goBlue disconnect 1            # Disconnect device #1

# Or use short names / substrings:
$ goBlue volume buds 75%         # Matches "realme Buds T200 Lite"
$ goBlue connect sony            # Matches "Sony WH-1000XM5"
$ goBlue info e8:d6              # Matches MAC suffix
```

### 1. Simple Connection & Direct Music Playback
Use this when you want to connect and stream music to a Bluetooth speaker/headphones **without** altering your permanent trust settings or changing the system's default audio sink:

```bash
# 1. Connect to the device (does NOT trust permanently or change system default)
$ goBlue connect "Sony WH-1000XM5"

# 2. Check current volume and adjust
$ goBlue audio volume "Sony WH-1000XM5"
Volume: 60% [unmuted]

$ goBlue audio volume "Sony WH-1000XM5" 80%
✓ Volume set to 80% for Sony WH-1000XM5

# 3. Verify sound with a quick test chime
$ goBlue audio play-test "Sony WH-1000XM5"
Playing test chime to Sony WH-1000XM5...
✓ Playback complete

# 4. Stream an audio file directly to the headphones
$ goBlue audio play "Sony WH-1000XM5" track.wav
Playing track.wav to Sony WH-1000XM5...
✓ Playback complete
```

---

### 2. Full System-Wide Bluetooth Audio Setup
Use this when you want your browser, media players, and desktop applications to permanently route sound to your Bluetooth device:

```bash
# 1. Scan for nearby devices
$ goBlue scan --timeout 8s

# 2. Pair and trust for automatic reconnects
$ goBlue pair "Sony WH-1000XM5"
$ goBlue trust "Sony WH-1000XM5"

# 3. Connect and set as default audio output
$ goBlue audio set-default "Sony WH-1000XM5"
✓ Device connected
✓ Bluetooth audio sink available
✓ Default audio output changed

# 4. Check status
$ goBlue audio status
```

---

## Command Reference

### 1. Adapter & Discovery

#### `goBlue adapter`
Displays the local Bluetooth adapter details.
```bash
$ goBlue adapter

BLUETOOTH ADAPTER

Adapter:       hci0
Address:       38:7A:0E:12:20:7C
Name:          mint-external
Powered:       yes
Discoverable:  no
Pairable:      yes
Discovering:   no
```

Specify a particular adapter using `--adapter`:
```bash
$ goBlue --adapter hci1 adapter
```

#### `goBlue adapter set-name <name>`
Changes the Bluetooth name (alias) broadcast to nearby discovering/connecting devices.
```bash
$ goBlue adapter set-name "My Custom Laptop"
✓ Adapter name set to "My Custom Laptop"
```

Reset back to system hostname:
```bash
$ goBlue adapter set-name --reset
✓ Adapter name reset to system default ("mint-external")
```

#### `goBlue devices`
Lists all known/paired devices stored in BlueZ.
```bash
$ goBlue devices

KNOWN BLUETOOTH DEVICES

NAME                     ADDRESS              PAIRED    TRUSTED    CONNECTED
realme Buds T200 Lite    28:04:C6:73:E8:D6    yes       yes        no
Sony WH-1000XM5          AA:BB:CC:DD:EE:FF    yes       yes        yes
```

#### `goBlue scan`
Performs an active discovery scan over the adapter for nearby devices.
```bash
$ goBlue scan --timeout 10s

Scanning for Bluetooth devices on hci0 (10s)...

BLUETOOTH DEVICES

NAME                     ADDRESS              RSSI    STATE
Sony WH-1000XM5          AA:BB:CC:DD:EE:FF    -48     paired
realme Buds T200 Lite    28:04:C6:73:E8:D6    -62     paired
Bluetooth Speaker        22:33:44:55:66:77    -75     available
```

#### `goBlue status`
Provides a comprehensive overview of the adapter and all currently connected devices with audio profiles.
```bash
$ goBlue status

BLUETOOTH STATUS

Adapter:      hci0
Powered:      yes

Connected devices:

Sony WH-1000XM5
    Address:   AA:BB:CC:DD:EE:FF
    Paired:    yes
    Trusted:   yes
    Connected: yes
    Audio:     yes
    Profile:   A2DP
    Default:   yes
```

---

### 2. Device Inspection & Audio Capabilities

#### `goBlue info <device>`
Inspects a device by its MAC address or exact Name/Alias.
```bash
$ goBlue info "realme Buds T200 Lite"

DEVICE

Name:     realme Buds T200 Lite
Address:  28:04:C6:73:E8:D6
Type:     audio-headset

Paired:     yes
Trusted:    yes
Connected:  no

AUDIO CAPABILITIES

A2DP Sink:    yes
A2DP Source:  no
HFP/HSP:      yes
AVRCP:        yes
```

#### `goBlue audio info <device>`
Performs a targeted audio role capability check to determine if the device can receive audio output from this Linux PC.
```bash
$ goBlue audio info "realme Buds T200 Lite"

AUDIO CAPABILITIES

Device:   realme Buds T200 Lite
Address:  28:04:C6:73:E8:D6

A2DP Sink:    yes
A2DP Source:  no
HFP/HSP:      yes
AVRCP:        yes

Can receive audio from this Linux machine:  yes
```

---

### 3. Device Connection & Management

#### `goBlue connect <device>`
Connects to a device. Does **not** set permanent trust and does **not** change system default audio routing.
```bash
$ goBlue connect "realme Buds T200 Lite"

Device found.
A2DP Sink detected.
Connecting...
✓ Connected
```

#### `goBlue disconnect <device>`
Disconnects from an active device.
```bash
$ goBlue disconnect "realme Buds T200 Lite"

✓ Disconnected
```

#### `goBlue pair <device>`
Pairs with a device via BlueZ D-Bus with transient agent support.
```bash
$ goBlue pair 28:04:C6:73:E8:D6

Device found
Pairing...
✓ Paired successfully
```

#### `goBlue trust <device>` / `goBlue untrust <device>`
Configures device trust in BlueZ to allow automatic reconnects.
```bash
$ goBlue trust "realme Buds T200 Lite"
✓ Trusted 28:04:C6:73:E8:D6

$ goBlue untrust "realme Buds T200 Lite"
✓ Untrusted 28:04:C6:73:E8:D6
```

---

### 4. Audio Control, Volume & Playback

#### `goBlue audio status`
Shows all Bluetooth audio devices, their connection state, active audio profile (A2DP / HFP / HSP), and default output marker.
```bash
$ goBlue audio status

BLUETOOTH AUDIO

DEVICE                   STATE           PROFILE    DEFAULT
realme Buds T200 Lite    connected       A2DP       yes
Galaxy Buds              disconnected    -          -
```

#### `goBlue audio volume <device> [level]`
Inspects or adjusts the volume level for the Bluetooth audio output.
```bash
# Check current volume
$ goBlue audio volume "realme Buds T200 Lite"
Volume: 75% [unmuted]

# Set specific volume level
$ goBlue audio volume "realme Buds T200 Lite" 80%
✓ Volume set to 80% for realme Buds T200 Lite

# Relative adjustments
$ goBlue audio volume "realme Buds T200 Lite" +10%
$ goBlue audio volume "realme Buds T200 Lite" -5%
```

#### `goBlue audio mute <device>` / `goBlue audio unmute <device>`
Mutes or unmutes the audio output on the device.
```bash
$ goBlue audio mute "realme Buds T200 Lite"
✓ Muted realme Buds T200 Lite

$ goBlue audio unmute "realme Buds T200 Lite"
✓ Unmuted realme Buds T200 Lite
```

#### `goBlue audio play-test <device>`
Generates and streams an acoustic stereo test chime directly through the device's PipeWire audio sink to verify sound.
```bash
$ goBlue audio play-test "realme Buds T200 Lite"
Playing test chime to realme Buds T200 Lite...
✓ Playback complete
```

#### `goBlue audio play <device> <file>`
Directly streams any audio file (WAV/FLAC/MP3) to the Bluetooth device without requiring it to be the system default.
```bash
$ goBlue audio play "realme Buds T200 Lite" music.wav
Playing music.wav to realme Buds T200 Lite...
✓ Playback complete
```

#### `goBlue audio set-default <device>`
Ensures the device is connected, verifies the PipeWire audio sink node, and sets it as the system's default audio sink.
```bash
$ goBlue audio set-default "realme Buds T200 Lite"

✓ Device connected
✓ Bluetooth audio sink available
✓ Default audio output changed
```

---

### 5. Audio Roles: A2DP Sink vs Source

- **A2DP Sink**: The device receives high-quality stereo audio (e.g. Headphones, Bluetooth Speakers, TVs acting as audio output).
- **A2DP Source**: The device sends high-quality stereo audio (e.g. Linux PC, Smartphone).

`goBlue` inspects standard Bluetooth service UUIDs (`0000110b-...` for A2DP Sink) rather than guessing capabilities from device names or icons.

---

### 6. JSON Output

Pass `--json` to any query command for structured automation output:

```bash
$ goBlue devices --json
$ goBlue status --json
$ goBlue audio status --json
$ goBlue audio volume "realme Buds T200 Lite" --json
$ goBlue audio info "realme Buds T200 Lite" --json
```
