package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	logitech "github.com/michaelabon/streamdeck-logitech-litra/internal/logitech_hid"
	"github.com/samwho/streamdeck"
	"github.com/sstallion/go-hid"
)

type Settings struct {
	Temperature uint16 `json:"temperature,string"`
	Brightness  uint8  `json:"brightness,string"`
}

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	fileName := "streamdeck-logitech-litra-lights.log"
	f, err := os.CreateTemp("logs", fileName)
	if err != nil {
		// The logger isn't set up yet, so stderr is the only channel left.
		fmt.Fprintf(os.Stderr, "unable to create log file %q: %v\n", fileName, err)
		exitCode = 83

		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	logLevel := slog.LevelInfo
	if os.Getenv("LITRA_DEBUG") != "" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("starting plugin",
		"debug_enabled", logLevel == slog.LevelDebug)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("fatal error", "error", err)
		exitCode = 1

		return
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

	turnOffLightsAction.RegisterHandler(
		streamdeck.KeyDown,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
			return handleTurnOffLights(ctx, client)
		},
	)
}

func setupSetLightsAction(client *streamdeck.Client, settings map[string]*Settings) {
	setLightsAction := client.Action("ca.michaelabon.logitech-litra-lights.set")

	setLightsAction.RegisterHandler(
		streamdeck.WillAppear,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
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

			if s.Temperature == 0 {
				s.Temperature = 3200
				s.Brightness = 50
			}

			background, err := streamdeck.Image(generateBackground(*s))
			if err != nil {
				slog.Error("failed to generate streamdeck image", "error", err)

				return err
			}

			if err := client.SetImage(ctx, background, streamdeck.HardwareAndSoftware); err != nil {
				return err
			}

			err = client.SetTitle(
				ctx,
				strconv.Itoa(int(s.Temperature)),
				streamdeck.HardwareAndSoftware,
			)

			return err
		},
	)

	setLightsAction.RegisterHandler(
		streamdeck.DidReceiveSettings,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
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
				slog.Error("failed to generate streamdeck image", "error", err)

				return err
			}

			if err := client.SetImage(ctx, background, streamdeck.HardwareAndSoftware); err != nil {
				return err
			}

			err = client.SetTitle(
				ctx,
				strconv.Itoa(int(s.Temperature)),
				streamdeck.HardwareAndSoftware,
			)

			return err
		},
	)

	setLightsAction.RegisterHandler(
		streamdeck.WillDisappear,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
			s := settings[event.Context]

			return client.SetSettings(ctx, s)
		},
	)

	setLightsAction.RegisterHandler(
		streamdeck.KeyDown,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
			return handleSetLights(ctx, client, event, settings)
		},
	)
}

func handleTurnOffLights(ctx context.Context, client *streamdeck.Client) error {
	slog.Debug("handleTurnOffLights called")

	err := writeToLights(sendTurnOffLights())
	if err != nil {
		slog.Error("failed to turn off lights", "error", err)

		return client.SetTitle(ctx, "Err", streamdeck.HardwareAndSoftware)
	}

	slog.Debug("handleTurnOffLights completed successfully")

	return nil
}

func handleSetLights(
	ctx context.Context,
	client *streamdeck.Client,
	event streamdeck.Event,
	settings map[string]*Settings,
) error {
	s, ok := settings[event.Context]
	if !ok {
		return fmt.Errorf("couldn't find settings for context %v", event.Context)
	}

	slog.Debug("handleSetLights called",
		"temperature", s.Temperature,
		"brightness", s.Brightness)

	if err := client.SetSettings(ctx, s); err != nil {
		return err
	}

	background, err := streamdeck.Image(generateBackground(*s))
	if err != nil {
		slog.Error("failed to generate streamdeck image", "error", err)

		return err
	}

	err = writeToLights(sendBrightnessAndTemperature(*s))
	if err != nil {
		slog.Error("failed to set lights", "error", err)

		return client.SetTitle(ctx, "Err", streamdeck.HardwareAndSoftware)
	}

	err = client.SetImage(ctx, background, streamdeck.HardwareAndSoftware)
	if err != nil {
		slog.Error("failed to set streamdeck image", "error", err)

		return err
	}

	slog.Debug("handleSetLights completed successfully")

	return client.SetTitle(ctx, strconv.Itoa(int(s.Temperature)), streamdeck.HardwareAndSoftware)
}

const (
	// VID is the USB Vendor ID for Logitech.
	VID = 0x046d

	// PID is the USB Product ID for the Litra Glow.
	// Other Litra products:
	// Beam = 0xc901, Beam LX = 0xc903.
	PID = 0xc900

	// Litra exposes multiple HID interfaces.
	// This one accepts brightness/temperature commands.
	// See DEVELOPERS.md for details.
	UsagePage = 0xff43
)

// writeToLights opens a connection to each light attached to the computer
// and then invokes theFunc for each light.
func writeToLights(theFunc hid.EnumFunc) error {
	slog.Debug("writeToLights: starting",
		"vid", fmt.Sprintf("0x%04x", VID),
		"pid", fmt.Sprintf("0x%04x", PID))

	if err := hid.Init(); err != nil {
		slog.Error("writeToLights: hid.Init() failed", "error", err)

		return err
	}
	defer func() {
		if err := hid.Exit(); err != nil {
			slog.Error("writeToLights: hid.Exit() failed", "error", err)
		}
	}()

	err := hid.Enumerate(VID, PID, theFunc)
	if err != nil {
		slog.Error("writeToLights: hid.Enumerate() failed", "error", err)

		return err
	}

	slog.Debug("writeToLights: completed successfully")

	return nil
}

func sendBrightnessAndTemperature(settings Settings) hid.EnumFunc {
	return func(deviceInfo *hid.DeviceInfo) error {
		if deviceInfo.UsagePage != UsagePage {
			slog.Debug("skipping non-control interface",
				"serial", deviceInfo.SerialNbr,
				"usagePage", fmt.Sprintf("0x%04x", deviceInfo.UsagePage),
				"expected", fmt.Sprintf("0x%04x", UsagePage))

			return nil
		}

		slog.Debug("found Litra control interface",
			"serial", deviceInfo.SerialNbr,
			"path", deviceInfo.Path,
			"usagePage", fmt.Sprintf("0x%04x", deviceInfo.UsagePage))

		d, err := hid.OpenPath(deviceInfo.Path)
		if err != nil {
			slog.Error("failed to open device",
				"serial", deviceInfo.SerialNbr,
				"error", err)

			return err
		}
		defer func(d *hid.Device) {
			if err := d.Close(); err != nil {
				slog.Error("failed to close device", "error", err)
			}
		}(d)

		slog.Debug("sending lights on command", "serial", deviceInfo.SerialNbr)
		byteSequence := logitech.ConvertLightsOn()
		if _, err := d.Write(byteSequence); err != nil {
			slog.Error("failed to send lights on command", "error", err)

			return err
		}

		slog.Debug("sending brightness command",
			"serial", deviceInfo.SerialNbr,
			"brightness", settings.Brightness)
		byteSequence, err = logitech.ConvertBrightness(settings.Brightness)
		if err != nil {
			slog.Error("failed to convert brightness", "error", err)

			return err
		}
		if _, err := d.Write(byteSequence); err != nil {
			slog.Error("failed to send brightness command", "error", err)

			return err
		}

		slog.Debug("sending temperature command",
			"serial", deviceInfo.SerialNbr,
			"temperature", settings.Temperature)
		byteSequence, err = logitech.ConvertTemperature(settings.Temperature)
		if err != nil {
			slog.Error("failed to convert temperature", "error", err)

			return err
		}
		if _, err := d.Write(byteSequence); err != nil {
			slog.Error("failed to send temperature command", "error", err)

			return err
		}

		slog.Debug("successfully set brightness and temperature",
			"serial", deviceInfo.SerialNbr)

		return nil
	}
}

func sendTurnOffLights() hid.EnumFunc {
	return func(deviceInfo *hid.DeviceInfo) error {
		if deviceInfo.UsagePage != UsagePage {
			slog.Debug("skipping non-control interface",
				"serial", deviceInfo.SerialNbr,
				"usagePage", fmt.Sprintf("0x%04x", deviceInfo.UsagePage),
				"expected", fmt.Sprintf("0x%04x", UsagePage))

			return nil
		}

		slog.Debug("found Litra control interface",
			"serial", deviceInfo.SerialNbr,
			"path", deviceInfo.Path,
			"usagePage", fmt.Sprintf("0x%04x", deviceInfo.UsagePage))

		d, err := hid.OpenPath(deviceInfo.Path)
		if err != nil {
			slog.Error("failed to open device",
				"serial", deviceInfo.SerialNbr,
				"error", err)

			return err
		}
		defer func(d *hid.Device) {
			if err := d.Close(); err != nil {
				slog.Error("failed to close device", "error", err)
			}
		}(d)

		slog.Debug("sending lights off command", "serial", deviceInfo.SerialNbr)
		byteSequence := logitech.ConvertLightsOff()
		if _, err := d.Write(byteSequence); err != nil {
			slog.Error("failed to send lights off command", "error", err)

			return err
		}

		slog.Debug("successfully turned off lights", "serial", deviceInfo.SerialNbr)

		return nil
	}
}
