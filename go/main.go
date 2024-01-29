package main

import (
	"context"
	"encoding/json"
	"fmt"
	logitech "github.com/michaelabon/streamdeck-logitech-litra/internal/logitech_hid"
	"github.com/samwho/streamdeck"
	"github.com/sstallion/go-hid"
	"log"
	"os"
	"strconv"
)

type Settings struct {
	Temperature uint16 `json:"temperature,string"`
	Brightness  uint8  `json:"brightness,string"`
}

func main() {
	fileName := "streamdeck-logitech-litra-lights.log"
	f, err := os.CreateTemp("logs", fileName)
	if err != nil {
		log.Fatalf("error creating temp file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("unable to close file “%s”: %v\n", fileName, err)
		}
	}(f)

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
	settings := make(map[string]*Settings)

	setupSetLightsAction(client, settings)
	setupTurnOffLightsAction(client)
}

func setupTurnOffLightsAction(client *streamdeck.Client) {
	turnOffLightsAction := client.Action("ca.michaelabon.logitech-litra-lights.off")

	turnOffLightsAction.RegisterHandler(streamdeck.KeyDown, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		return handleTurnOffLights(ctx, client)
	})
}

func setupSetLightsAction(client *streamdeck.Client, settings map[string]*Settings) {
	setLightsAction := client.Action("ca.michaelabon.logitech-litra-lights.set")

	setLightsAction.RegisterHandler(streamdeck.WillAppear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
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

	setLightsAction.RegisterHandler(streamdeck.DidReceiveSettings, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
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

	setLightsAction.RegisterHandler(streamdeck.WillDisappear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		s, _ := settings[event.Context]
		return client.SetSettings(ctx, s)
	})

	setLightsAction.RegisterHandler(streamdeck.KeyDown, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		return handleSetLights(ctx, client, event, settings)
	})
}

func handleTurnOffLights(ctx context.Context, client *streamdeck.Client) error {
	err := writeToLights(sendTurnOffLights())
	if err != nil {
		log.Println("Error: ", err)
		return client.SetTitle(ctx, "Err", streamdeck.HardwareAndSoftware)
	}

	return nil
}

func handleSetLights(ctx context.Context, client *streamdeck.Client, event streamdeck.Event, settings map[string]*Settings) error {
	s, ok := settings[event.Context]
	if !ok {
		return fmt.Errorf("couldn't find settings for context %v", event.Context)
	}

	log.Printf("KeyDown with payload %+v\n", event.Payload)

	if err := client.SetSettings(ctx, s); err != nil {
		return err
	}

	background, err := streamdeck.Image(generateBackground(*s))
	if err != nil {
		log.Println("Error while generating streamdeck image", err)
		return err
	}

	err = writeToLights(sendBrightnessAndTemperature(*s))
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
}

const VID = 0x046d
const PID = 0xc900

// writeToLights opens a connection to each light attached to the computer
// and then invokes theFunc for each light.
func writeToLights(theFunc hid.EnumFunc) error {
	var err error

	if err = hid.Init(); err != nil {
		log.Println("Unable to hid.Init()", err)
		log.Println(err)
	}
	defer func() {
		err := hid.Exit()
		if err != nil {
			log.Println("unable to hid.Exit()", err)
		}
	}()

	err = hid.Enumerate(VID, PID, theFunc)
	if err != nil {
		return err
	}

	return nil
}

func sendBrightnessAndTemperature(settings Settings) hid.EnumFunc {
	return func(deviceInfo *hid.DeviceInfo) error {
		var err error

		d, err := hid.Open(VID, PID, deviceInfo.SerialNbr)
		if err != nil {
			log.Println("Unable to open", err)
			return err
		}
		defer func(d *hid.Device) {
			err := d.Close()
			if err != nil {
				log.Println("unable to hid.Device.Close()", err)
			}
		}(d)

		byteSequence := logitech.ConvertLightsOn()
		if _, err := d.Write(byteSequence); err != nil {
			log.Println(err)
			return err
		}

		byteSequence, err = logitech.ConvertBrightness(settings.Brightness)
		if err != nil {
			log.Println(err)
			return err
		}
		if _, err := d.Write(byteSequence); err != nil {
			log.Println("Unable to write bytes with set brightness", err)
			return err
		}
		byteSequence, err = logitech.ConvertTemperature(settings.Temperature)
		if err != nil {
			log.Println(err)
			return err
		}
		if _, err := d.Write(byteSequence); err != nil {
			log.Println("Unable to write bytes with set temperature", err)
			return err
		}

		return nil
	}
}

func sendTurnOffLights() hid.EnumFunc {
	return func(deviceInfo *hid.DeviceInfo) error {
		var err error

		d, err := hid.Open(VID, PID, deviceInfo.SerialNbr)
		if err != nil {
			log.Println("unable to open", err)
			return err
		}
		defer func(d *hid.Device) {
			err := d.Close()
			if err != nil {
				log.Println("unable to hid.Device.Close()", err)
			}
		}(d)

		byteSequence := logitech.ConvertLightsOff()
		if _, err := d.Write(byteSequence); err != nil {
			log.Println("unable to write bytes with lights off", err)
			return err
		}

		return nil
	}
}
