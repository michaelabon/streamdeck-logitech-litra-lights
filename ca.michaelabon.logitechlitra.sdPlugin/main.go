package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sstallion/go-hid"
	"image"
	"image/color"
	"log"
	"os"
	"strconv"

	tt "github.com/maruel/temperature"
	"github.com/samwho/streamdeck"
)

import "golang.org/x/exp/constraints"

type Settings struct {
	Counter     int    `json:"counter"`
	Temperature uint16 `json:"temperature,string"`
	Brightness  uint8  `json:"brightness,string"`
}

func main() {
	f, err := os.CreateTemp("logs", "streamdeck-counter.log")
	if err != nil {
		log.Fatalf("error creating temp file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func run(ctx context.Context) error {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return err
	}

	client := streamdeck.NewClient(ctx, params)
	setup(client)

	return client.Run()
}

func setup(client *streamdeck.Client) {
	action := client.Action("dev.samwho.streamdeck.counter")
	settings := make(map[string]*Settings)

	action.RegisterHandler(streamdeck.WillAppear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p := streamdeck.WillAppearPayload{}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		s, ok := settings[event.Context]
		if !ok {
			s = &Settings{}
			settings[event.Context] = s
		}

		if err := json.Unmarshal(p.Settings, s); err != nil {
			return err
		}

		background, err := streamdeck.Image(generateBackground(*s))
		if err != nil {
			log.Println("Error while generating streamdeck image", err)
			return err
		}

		if err := client.SetImage(ctx, background, streamdeck.HardwareAndSoftware); err != nil {
			return err
		}

		err = client.SetTitle(ctx, strconv.Itoa(int(s.Temperature)), streamdeck.HardwareAndSoftware)

		return err
	})

	action.RegisterHandler(streamdeck.DidReceiveSettings, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p := streamdeck.DidReceiveSettingsPayload{}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		s, ok := settings[event.Context]
		if !ok {
			s = &Settings{}
			settings[event.Context] = s
		}

		if err := json.Unmarshal(p.Settings, s); err != nil {
			return err
		}

		background, err := streamdeck.Image(generateBackground(*s))
		if err != nil {
			log.Println("Error while generating streamdeck image", err)
			return err
		}

		if err := client.SetImage(ctx, background, streamdeck.HardwareAndSoftware); err != nil {
			return err
		}

		err = client.SetTitle(ctx, strconv.Itoa(int(s.Temperature)), streamdeck.HardwareAndSoftware)

		return err
	})

	action.RegisterHandler(streamdeck.WillDisappear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		s, _ := settings[event.Context]
		s.Counter = 0
		return client.SetSettings(ctx, s)
	})

	action.RegisterHandler(streamdeck.KeyDown, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		s, ok := settings[event.Context]
		if !ok {
			return fmt.Errorf("couldn't find settings for context %v", event.Context)
		}

		log.Printf("KeyDown with payload %+v\n", event.Payload)

		s.Counter++
		if err := client.SetSettings(ctx, s); err != nil {
			return err
		}

		background, err := streamdeck.Image(generateBackground(*s))
		if err != nil {
			log.Println("Error while generating streamdeck image", err)
			return err
		}

		err = invokeLights(*s)
		if err != nil {
			log.Println("Error: ", err)
			return client.SetTitle(ctx, "Err", streamdeck.HardwareAndSoftware)
		}

		err = client.SetImage(ctx, background, streamdeck.HardwareAndSoftware)
		if err != nil {
			log.Println("Error while setting the light background", err)
			return err
		}

		return client.SetTitle(ctx, strconv.Itoa(int(s.Temperature)), streamdeck.HardwareAndSoftware)
	})
}

const VID = 0x046d
const PID = 0xc900

func lightsOn() (b []byte) {
	b = make([]byte, 20)
	copy(b, []byte{0x11, 0xff, 0x04, 0x1c, 0x01})
	return
}

func lightsOff() (b []byte) {
	b = make([]byte, 20)
	copy(b, []byte{0x11, 0xff, 0x04, 0x1c, 0x00})
	return
}

func setBrightness(percentage uint8) (b []byte) {
	b = make([]byte, 20)
	copy(b, []byte{0x11, 0xff, 0x04, 0x4c, 0x00, calcBrightness(percentage)})
	return
}

// Takes 1-100 and returns 20-250
func calcBrightness(brightness uint8) byte {
	return byte(int(float64(brightness-1.0)/(99.0)*(250-20)) + 20)
}

// Takes 2700-6500
func setTemperature(temperature uint16) (b []byte) {
	b = make([]byte, 20)
	b[0] = 0x11
	b[1] = 0xff
	b[2] = 0x04
	b[3] = 0x9c
	b[4] = byte(temperature >> 8)
	b[5] = byte(temperature)
	return
}

func invokeLights(settings Settings) error {
	var err error

	if err = hid.Init(); err != nil {
		log.Println("Unable to hid.Init()", err)
		log.Println(err)
	}

	err = hid.Enumerate(VID, PID, func(info *hid.DeviceInfo) error {
		var err error

		d, err := hid.Open(VID, PID, info.SerialNbr)
		if err != nil {
			log.Println("Unable to open", err)
			return err
		}
		defer d.Close()

		bb := lightsOn()
		if _, err := d.Write(bb); err != nil {
			log.Println("Unable to write bytes with lights on", err)
			return err
		}
		bb = setBrightness(settings.Brightness)
		if _, err := d.Write(bb); err != nil {
			log.Println("Unable to write bytes with set brightness", err)
			return err
		}
		bb = setTemperature(settings.Temperature)
		if _, err := d.Write(bb); err != nil {
			log.Println("Unable to write bytes with set temperature", err)
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Finalize the hid package.
	if err := hid.Exit(); err != nil {
		return err
	}

	return nil
}

func blackBackground() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 72, 72))
	for x := 0; x < 72; x++ {
		for y := 0; y < 72; y++ {
			img.Set(x, y, color.Black)
		}
	}
	return img
}

func clamp[K constraints.Ordered](n K, min, max K) K {
	r := n
	if r < min {
		r = min
	}
	if r > max {
		r = max
	}
	return r
}

func clampBrightness(n int8) int8 {
	return clamp(n, 0, 100)
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

	r, g, b := tt.ToRGB(temperature)

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

// Returns the interpolated value the is calculated from topC to botC
func lerp(topC, botC uint8, min, y, max int) uint8 {
	y = clamp(y, min, max)
	percentage := float64(y-min) / float64(max-min)
	value := topC - uint8(float64(topC-botC)*percentage)

	return value
}
