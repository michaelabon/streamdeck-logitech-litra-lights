# Developer Documentation

This document contains technical details about how the Logitech Litra lights communicate over USB HID,
and about the project

## Table of Contents

::toc::

---

## HID Basics

### What is HID?

**HID (Human Interface Device)** is a USB protocol originally designed for keyboards, mice, and joysticks.
It's now used for many devices because:

- No custom drivers required - the OS has built-in support
- Simple, well-documented protocol
- Works across Windows, macOS, and Linux

### Interfaces

A single USB device can expose **multiple logical interfaces**.
Think of it like a Swiss Army knife - one physical device, multiple tools.
Each interface has its own "path" on the system and serves a different purpose.

### Usage Pages

Usage Pages categorize what a device interface does.
They're defined by the [USB HID specification](https://usb.org/hid):

| Usage Page        | Meaning                                     |
| ----------        | -------                                     |
| `0x0001`          | Generic Desktop (mouse, keyboard, joystick) |
| `0x0007`          | Keyboard/Keypad                             |
| `0x000C`          | Consumer (media controls)                   |
| `0xFF00`-`0xFFFF` | **Vendor-defined** (custom protocols)       |

Logitech uses `0xFF43` for their proprietary lighting control protocol.

---

## Logitech Litra HID Details

### Supported Devices

| Device        | Vendor ID | Product ID | Status    |
| ------        | --------- | ---------- | ------    |
| Litra Glow    | `0x046d`  | `0xc900`   | Supported |
| Litra Beam    | `0x046d`  | `0xc901`   | Untested  |
| Litra Beam LX | `0x046d`  | `0xc903`   | Untested  |

### HID Interfaces per Device

Each Litra device exposes **two HID interfaces**:

| Interface | Usage Page | Purpose                                    |
| --------: | ---------: | -------                                    |
| 1         | `0x000C`   | Consumer Control (do not use for commands) |
| 2         | `0xFF43`   | **Vendor-specific (commands go here!)**    |

Both interfaces share the same Usage value: `0x0202`. I don't know what that means.

### Example Enumeration Output

```
Product: Litra Glow
  PID:        0xc900
  Path:       DevSrvsID:4295441247
  UsagePage:  0x000c        // <-- Wrong interface (Consumer Control)
  Usage:      0x0001
  Serial:     2229FE000MK8

Product: Litra Glow
  PID:        0xc900
  Path:       DevSrvsID:4295441247
  UsagePage:  0xff43        // <-- Correct interface (Vendor-specific)
  Usage:      0x0202
  Serial:     2229FE000MK8
```

## HID Protocol

All commands are **20 bytes**, padded with zeros.

### Output Commands (Computer → Device)

| Command         | Bytes               | Description                                |
| -------         | -----               | -----------                                |
| Power On        | `11 ff 04 1c 01`    | Turn light on                              |
| Power Off       | `11 ff 04 1c 00`    | Turn light off                             |
| Set Brightness  | `11 ff 04 4c 00 XX` | XX = brightness (0x14-0xFA, i.e., 20-250)  |
| Set Temperature | `11 ff 04 9c HH LL` | HH:LL = temperature in Kelvin (big-endian) |

### Brightness Conversion

The device uses lumen values 20-250, not percentages:

```go
// Percentage (1-100) to device units (20-250)
func percentToDevice(percent uint8) byte {
    return byte((float64(percent-1) / 99.0 * 230) + 20)
}

// Device units (20-250) to percentage (1-100)
func deviceToPercent(device byte) int {
    return int((float64(device)-20)/(250-20)*99) + 1
}
```

I don't get it either.

### Temperature Range

- Minimum: 2700K (warm white)
- Maximum: 6500K (cool white)
- Step: 100K increments

```go
// Temperature is sent as big-endian uint16
tempBytes := []byte{byte(temp >> 8), byte(temp & 0xFF)}
```

---

## Reading Device State

I don't know how to read events from the Litra **does not push events** when physical buttons are pressed.
The official Logitech G HUB app reads pretty instantaneously, updating its UI, but I don't know how to do that.

To get current state, AFAIK, you must query the device.

### Query Commands

| Query       | Command       | Response                        |
| -----       | -------       | --------                        |
| Brightness  | `11 ff 04 31` | Byte 5 = brightness (20-250)    |
| Temperature | `11 ff 04 81` | Bytes 4-5 = Kelvin (big-endian) |
| Power State | `11 ff 04 01` | Byte 4 = 1 (on) or 0 (off)      |

### Query Protocol

1. Write the query command (20 bytes)
2. Read the response (20 bytes)
3. Parse the relevant bytes

### Important: Read Behavior

- **Blocking `Read()`** works correctly
- **`ReadWithTimeout()`** may not work on macOS - use a goroutine with timeout instead

```go
func queryWithTimeout(device *hid.Device, cmd []byte, timeout time.Duration) ([]byte, error) {
    _, err := device.Write(cmd)
    if err != nil {
        return nil, err
    }

    response := make([]byte, 20)
    done := make(chan error)

    go func() {
        _, err := device.Read(response)
        done <- err
    }()

    select {
    case err := <-done:
        return response, err
    case <-time.After(timeout):
        return nil, fmt.Errorf("timeout")
    }
}
```

### Example: Query Current State

```go
// Query brightness
brightnessCmd := make([]byte, 20)
copy(brightnessCmd, []byte{0x11, 0xff, 0x04, 0x31})
device.Write(brightnessCmd)
response := make([]byte, 20)
device.Read(response)
brightness := response[5]  // 20-250

// Query temperature
tempCmd := make([]byte, 20)
copy(tempCmd, []byte{0x11, 0xff, 0x04, 0x81})
device.Write(tempCmd)
device.Read(response)
temperature := int(response[4])*256 + int(response[5])  // 2700-6500
```

---

## Differentiating Multiple Devices

Each Litra has a unique **serial number** available in `DeviceInfo.SerialNbr`:

```go
hid.Enumerate(VID, PID, func(info *hid.DeviceInfo) error {
    if info.UsagePage == 0xff43 {
        fmt.Printf("Found Litra: %s\n", info.SerialNbr)
        // Serial numbers look like: 2229FE000MK8, 2229FE000M38
    }
    return nil
})
```

To control a specific device,
match the serial number when opening:

```go
hid.Enumerate(VID, PID, func(info *hid.DeviceInfo) error {
    if info.UsagePage == 0xff43 && info.SerialNbr == targetSerial {
        device, _ := hid.OpenPath(info.Path)
        // Control this specific device
    }
    return nil
})
```

---

## Cross-Platform Considerations

I own a Windows and a Mac computer. Other than WSL, I don't have a Linux desktop that I use regularly.

---

## Related Projects & References

### Other Litra Control Projects

- [timrogers/litra](https://github.com/timrogers/litra) - JavaScript/Node.js library with CLI
- [kharyam/litra-driver](https://github.com/kharyam/litra-driver) - Python driver with GUI
- [kharyam/go-litra-driver](https://github.com/kharyam/go-litra-driver) - Go driver (uses same go-hid library)

### HID Libraries

- [sstallion/go-hid](https://github.com/sstallion/go-hid) - Go bindings for hidapi (used by this project)
- [libusb/hidapi](https://github.com/libusb/hidapi) - Cross-platform HID library (C)

### Tools for Debugging

- [hidapitester](https://github.com/todbot/hidapitester) - Command-line HID testing tool

  ```bash
  # Query brightness
  hidapitester --vidpid 046D/C900 --usagePage FF43 --open \
    --length 20 --send-output 0x11,0xff,0x04,0x31 --read-input

  # Turn light on
  hidapitester --vidpid 046D/C900 --usagePage FF43 --open \
    --length 20 --send-output 0x11,0xff,0x04,0x1c,0x01
  ```


