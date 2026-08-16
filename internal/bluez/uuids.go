package bluez

import (
	"strings"
)

// Well-known Bluetooth SIG UUIDs.
const (
	UUIDA2DPSource             = "0000110a-0000-1000-8000-00805f9b34fb" // Audio Source (Sends audio)
	UUIDA2DPSink               = "0000110b-0000-1000-8000-00805f9b34fb" // Audio Sink (Receives audio)
	UUIDAVRCPTarget            = "0000110c-0000-1000-8000-00805f9b34fb" // AVRCP Target
	UUIDAVRCPController        = "0000110e-0000-1000-8000-00805f9b34fb" // AVRCP Controller
	UUIDHeadset                = "00001108-0000-1000-8000-00805f9b34fb" // Headset (HSP)
	UUIDHeadsetAudioGateway    = "00001112-0000-1000-8000-00805f9b34fb" // Headset Audio Gateway (HSP AG)
	UUIDHandsfree              = "0000111e-0000-1000-8000-00805f9b34fb" // Handsfree (HFP)
	UUIDHandsfreeAudioGateway  = "0000111f-0000-1000-8000-00805f9b34fb" // Handsfree Audio Gateway (HFP AG)
	UUIDHumanInterfaceDevice   = "00001124-0000-1000-8000-00805f9b34fb" // HID
	UUIDSerialPort             = "00001101-0000-1000-8000-00805f9b34fb" // SPP
	UUIDGenericAccess          = "00001800-0000-1000-8000-00805f9b34fb" // GAP
	UUIDGenericAttribute       = "00001801-0000-1000-8000-00805f9b34fb" // GATT
	UUIDDeviceInformation      = "0000180a-0000-1000-8000-00805f9b34fb" // DIS
	UUIDBatteryService         = "0000180f-0000-1000-8000-00805f9b34fb" // BAS
	UUIDPhonebookAccessServer  = "0000112f-0000-1000-8000-00805f9b34fb" // PBAP Server
	UUIDMessageAccessServer    = "00001132-0000-1000-8000-00805f9b34fb" // MAP Server
	UUIDMessageNotificationSrv = "00001133-0000-1000-8000-00805f9b34fb" // MAP Notification
	UUIDOBEXObjectPush         = "00001105-0000-1000-8000-00805f9b34fb" // OPP
	UUIDOBEXFileTransfer       = "00001106-0000-1000-8000-00805f9b34fb" // FTP
	UUIDPnPInformation         = "00001200-0000-1000-8000-00805f9b34fb" // PnP
)

var knownUUIDNames = map[string]string{
	UUIDA2DPSource:             "Audio Source (A2DP)",
	UUIDA2DPSink:               "Audio Sink (A2DP)",
	UUIDAVRCPTarget:            "A/V Remote Control Target",
	UUIDAVRCPController:        "A/V Remote Control",
	UUIDHeadset:                "Headset (HSP)",
	UUIDHeadsetAudioGateway:    "Headset Audio Gateway",
	UUIDHandsfree:              "Handsfree (HFP)",
	UUIDHandsfreeAudioGateway:  "Handsfree Audio Gateway",
	UUIDHumanInterfaceDevice:   "Human Interface Device",
	UUIDSerialPort:             "Serial Port",
	UUIDGenericAccess:          "Generic Access Profile",
	UUIDGenericAttribute:       "Generic Attribute Profile",
	UUIDDeviceInformation:      "Device Information",
	UUIDBatteryService:         "Battery Service",
	UUIDPhonebookAccessServer:  "Phonebook Access Server",
	UUIDMessageAccessServer:    "Message Access Server",
	UUIDMessageNotificationSrv: "Message Notification Server",
	UUIDOBEXObjectPush:         "OBEX Object Push",
	UUIDOBEXFileTransfer:       "OBEX File Transfer",
	UUIDPnPInformation:         "PnP Information",
}

// NormalizeUUID standardizes a UUID to lowercase 128-bit format.
func NormalizeUUID(uuid string) string {
	cleaned := strings.ToLower(strings.TrimSpace(uuid))
	// If 16-bit hex e.g. "110b" or "0x110b"
	cleaned = strings.TrimPrefix(cleaned, "0x")
	if len(cleaned) == 4 {
		return "0000" + cleaned + "-0000-1000-8000-00805f9b34fb"
	}
	if len(cleaned) == 8 {
		return cleaned + "-0000-1000-8000-00805f9b34fb"
	}
	return cleaned
}

// ServiceNameFromUUID returns a human-readable service name for a UUID.
func ServiceNameFromUUID(uuid string) string {
	norm := NormalizeUUID(uuid)
	if name, ok := knownUUIDNames[norm]; ok {
		return name
	}
	return norm
}
