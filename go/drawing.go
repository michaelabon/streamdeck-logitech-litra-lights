package main

import (
	"image"
	"image/color"
	"log"

	temperature2 "github.com/maruel/temperature"
)

func clamp(n, min, max int) int {
	r := n
	if r < min {
		r = min
	}
	if r > max {
		r = max
	}
	return r
}

func generateBackground(settings Settings) image.Image {
	const dim = 72 // TODO: Does this need to be 144?
	img := image.NewRGBA64(image.Rect(0, 0, dim, dim))

	temperature := settings.Temperature
	brightness := settings.Brightness

	// scale temperature so that blue looks bluer
	temperature = uint16(float64(temperature-2700)*1.3 + 2700)

	brightness = 100 - brightness
	brightness = uint8(float64(brightness) / 1.2)

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

			c := color.RGBA{R: r2, G: g2, B: b2, A: 0xff}
			img.Set(x, y, c)
		}
	}

	return img
}

// Returns the interpolated value that is calculated from topC to botC
func lerp(topC, botC uint8, min, y, max int) uint8 {
	y = clamp(y, min, max)
	percentage := float64(y-min) / float64(max-min)
	value := topC - uint8(float64(topC-botC)*percentage)

	return value
}
