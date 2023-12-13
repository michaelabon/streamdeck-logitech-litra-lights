package main

import "testing"

func Test_calcBrightness(t *testing.T) {
	tests := []struct {
		name       string
		brightness uint8
		want       byte
	}{
		{
			"1 -> 20",
			1,
			20,
		},
		{
			"100 -> 250",
			100,
			250,
		},
		{
			"50 -> 133",
			50,
			133,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcBrightness(tt.brightness); got != tt.want {
				t.Errorf("calcBrightness() = %v, want %v", got, tt.want)
			}
		})
	}
}
