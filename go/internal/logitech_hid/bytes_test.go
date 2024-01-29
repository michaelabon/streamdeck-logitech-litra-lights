package logitech_hid

import (
	"testing"
)

func TestConvertBrightness(t *testing.T) {
	tests := []struct {
		input    uint8
		expected []byte
		errMsg   string
	}{
		{1, extendByteSlice([]byte{0x11, 0xff, 0x04, 0x4c, 0x00, 0x14}), ""},
		{50, extendByteSlice([]byte{0x11, 0xff, 0x04, 0x4c, 0x00, 0x85}), ""},
		{100, extendByteSlice([]byte{0x11, 0xff, 0x04, 0x4c, 0x00, 0xfa}), ""},
		{0, nil, "percentage must be greater than 1, was 0"},
		{101, nil, "percentage must be less than 100, was 101"},
	}

	for _, test := range tests {
		result, err := ConvertBrightness(test.input)

		if err != nil && err.Error() != test.errMsg {
			t.Errorf("For input %d, expected error message '%s', but got '%s'", test.input, test.errMsg, err.Error())
		}

		if err == nil && len(result) != len(test.expected) {
			t.Errorf("For input %d, expected byte slice of length %d, but got length %d", test.input, len(test.expected), len(result))
		}

		if err == nil {
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("For input %d, at index %d, expected %x, but got %x", test.input, i, test.expected[i], result[i])
				}
			}
		}
	}
}

func TestConvertTemperature(t *testing.T) {
	tests := []struct {
		input    uint16
		expected []byte
		errMsg   string
	}{
		{2700, extendByteSlice([]byte{0x11, 0xff, 0x04, 0x9c, 0x0a, 0x8c}), ""},
		{4000, extendByteSlice([]byte{0x11, 0xff, 0x04, 0x9c, 0x0f, 0xa0}), ""},
		{6500, extendByteSlice([]byte{0x11, 0xff, 0x04, 0x9c, 0x19, 0x64}), ""},
		{2699, nil, "temperature must be greater than 2700, was 2699"},
		{6501, nil, "temperature must be less than 6500, was 6501"},
	}

	for _, test := range tests {
		result, err := ConvertTemperature(test.input)

		if err != nil && err.Error() != test.errMsg {
			t.Errorf("For input %d, expected error message '%s', but got '%s'", test.input, test.errMsg, err.Error())
		}

		if err == nil && len(result) != len(test.expected) {
			t.Errorf("For input %d, expected byte slice of length %d, but got length %d", test.input, len(test.expected), len(result))
		}

		if err == nil {
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("For input %d, at index %d, expected %x, but got %x", test.input, i, test.expected[i], result[i])
				}
			}
		}
	}
}

func TestCalcBrightness(t *testing.T) {
	tests := []struct {
		input    uint8
		expected byte
	}{
		{1, 20},
		{100, 250},
		{50, 133},
	}
	for _, test := range tests {
		result := calcBrightness(test.input)
		if result != test.expected {
			t.Errorf("For input %d, expected %x, but got %x", test.input, test.expected, result)
		}
	}
}

func extendByteSlice(a []byte) []byte {
	b := make([]byte, 20)
	copy(b, a)

	return b
}
