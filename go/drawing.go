package main

import (
	"image"
	"image/color"
	"log"

	temperatureconverter "github.com/maruel/temperature"
)

const (
	minTemperature = 2700
	maxBrightness  = 100
	opaque         = uint8(0xff)
)

// temperatureScale is hand-tuned using a real Stream Deck device
// so that the blue looks bluer when generating the background
const temperatureScale = 1.3

// brightnessScale is hand-tuned using a real Stream Deck device
// so that the background is a little dimmer
const brightnessScale = 1.2

// generateBackground generates a background image for the Stream Deck
//
// The design is a vertical gradient. The top of the image is the user-selected
// temperature at full brightness. The bottom of the image is the same colour,
// but with the brightness reduced by the user-selected amount.
// It can look a bit like a Web 2.0 gradient for low-brightness settings,
// since the top of the image is always at full brightness, but the bottom
// will be relatively dark.
func generateBackground(settings Settings) image.Image {
	const dim = 72 // TODO: Does this need to be 144?
	img := image.NewRGBA64(image.Rect(0, 0, dim, dim))

	temperature := settings.Temperature
	brightness := settings.Brightness

	// scale temperature so that blue looks bluer
	temperature = uint16(float64(temperature-minTemperature)*temperatureScale + minTemperature)
	r, g, b := temperatureconverter.ToRGB(temperature)
	brightness = maxBrightness - brightness
	brightness = uint8(float64(brightness) / brightnessScale)

	log.Printf("Initial rgb{%v, %v, %v}\n", r, g, b)

	var r1, g1, b1 uint8

	// scale for brightness
	if r < brightness {
		r1 = 0
	} else {
		r1 = r - brightness
	}

	if g < brightness {
		g1 = 0
	} else {
		g1 = g - brightness
	}

	if b < brightness {
		b1 = 0
	} else {
		b1 = b - brightness
	}

	log.Printf("Final rgb{%v, %v, %v}\n", r, g, b)

	// The first 20 pixels will be shown at full brightness
	alwaysShowTheFullColorForTheFirstXPixels := 20
	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			r2 := lerp(r, r1, alwaysShowTheFullColorForTheFirstXPixels, y, dim)
			g2 := lerp(g, g1, alwaysShowTheFullColorForTheFirstXPixels, y, dim)
			b2 := lerp(b, b1, alwaysShowTheFullColorForTheFirstXPixels, y, dim)

			c := color.RGBA{R: r2, G: g2, B: b2, A: opaque}
			img.Set(x, y, c)
		}
	}

	return img
}

// Returns the interpolated value that is calculated from topC to botC
//
// topC - the starting colour component value (original colour
// botC - the ending colour component value (brightness-adjusted colour)
// minY - the minimum bound of the range
// y - the current position being evaluated (y-coordinate in the image)
// maxY - the maximum bound of the range
func lerp(topC, botC uint8, minY, y, maxY int) uint8 {
	y = clamp(y, minY, maxY)
	percentage := float64(y-minY) / float64(maxY-minY)
	value := topC - uint8(float64(topC-botC)*percentage)

	return value
}

func clamp(n, minVal, maxVal int) int {
	if n < minVal {
		return minVal
	}
	if n > maxVal {
		return maxVal
	}

	return n
}
