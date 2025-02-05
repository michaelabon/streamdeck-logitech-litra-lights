package main

import (
	"image"
	"image/color"
	"log"

	temperature2 "github.com/maruel/temperature"
)

func clamp(n, minV, maxV int) int {
	r := n
	if r < minV {
		r = minV
	}
	if r > maxV {
		r = maxV
	}

	return r
}

const (
	minSupportedTemp       = 2700
	maxSupportedBrightness = 100
	brightnessScaleFactor  = 1.2 // Because 0 will be too dark
	fullOpacity            = 0xff
)

func generateBackground(settings Settings) image.Image {
	const dim = 72 // TODO: Does this need to be 144?
	img := image.NewRGBA64(image.Rect(0, 0, dim, dim))

	temperature := settings.Temperature
	brightness := settings.Brightness

	// scale temperature so that blue looks bluer
	temperature = uint16(float64(temperature-minSupportedTemp)*1.3 + minSupportedTemp)

	brightness = maxSupportedBrightness - brightness
	brightness = uint8(float64(brightness) / brightnessScaleFactor)

	r, g, b := temperature2.ToRGB(temperature)

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

	alwaysShowTheFullColorForTheFirstXPixels := 20
	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			r2 := lerp(r, r1, alwaysShowTheFullColorForTheFirstXPixels, y, dim)
			g2 := lerp(g, g1, alwaysShowTheFullColorForTheFirstXPixels, y, dim)
			b2 := lerp(b, b1, alwaysShowTheFullColorForTheFirstXPixels, y, dim)

			c := color.RGBA{R: r2, G: g2, B: b2, A: fullOpacity}
			img.Set(x, y, c)
		}
	}

	return img
}

// Returns the interpolated value that is calculated from topC to botC
func lerp(topC, botC uint8, minY, y, maxY int) uint8 {
	y = clamp(y, minY, maxY)
	percentage := float64(y-minY) / float64(maxY-minY)
	value := topC - uint8(float64(topC-botC)*percentage)

	return value
}
