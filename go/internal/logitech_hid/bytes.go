package logitech_hid

import (
	"fmt"
)

// Logitech expects to receive 20 bytes when given a command.
const ByteLength = 20

func ConvertLightsOn() (b []byte) {
	b = make([]byte, ByteLength)
	copy(b, []byte{0x11, 0xff, 0x04, 0x1c, 0x01})
	return
}

func ConvertLightsOff() (b []byte) {
	b = make([]byte, 20)
	copy(b, []byte{0x11, 0xff, 0x04, 0x1c, 0x00})
	return
}

// ConvertBrightness
//
// Expects 1-100
func ConvertBrightness(percentage uint8) ([]byte, error) {
	if percentage < 1 {
		return nil, fmt.Errorf("percentage must be greater than 1, was %d", percentage)
	}
	if percentage > 100 {
		return nil, fmt.Errorf("percentage must be less than 100, was %d", percentage)
	}

	b := make([]byte, ByteLength)
	copy(b, []byte{0x11, 0xff, 0x04, 0x4c, 0x00, calcBrightness(percentage)})
	return b, nil
}

// ConvertTemperature
//
// Expects 2700-6500
func ConvertTemperature(temperature uint16) ([]byte, error) {
	if temperature < 2700 {
		return nil, fmt.Errorf("temperature must be greater than 2700, was %d", temperature)
	}
	if temperature > 6500 {
		return nil, fmt.Errorf("temperature must be less than 6500, was %d", temperature)
	}

	b := make([]byte, ByteLength)
	b[0] = 0x11
	b[1] = 0xff
	b[2] = 0x04
	b[3] = 0x9c
	b[4] = byte(temperature >> 8)
	b[5] = byte(temperature)
	return b, nil
}

// Takes 1-100 and returns 20-250
//
// For some reason, the Logitech HID API expects to receive
//
//	    1% brightness as the byte  20 or 0x14, and
//		 100% brightness as the byte 250 or 0xfa,
//		 (and everything in between)
func calcBrightness(brightness uint8) byte {
	return byte(int(float64(brightness-1.0)/(99.0)*(250-20)) + 20)
}
