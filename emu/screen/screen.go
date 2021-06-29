package screen

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Window struct {
	*pixelgl.Window
	KeyMap     map[uint16]pixelgl.Button
	KeysPushed [16]*time.Ticker
}

func NewScreen() (*Window, error) {
	cfg := pixelgl.WindowConfig{
		Title:  "Chyp8",
		VSync:  true,
		Bounds: pixel.R(0, 0, 512, 384),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return nil, fmt.Errorf("Error when making new window: %v", err)
	}

	km := map[uint16]pixelgl.Button{
		0x01: pixelgl.Key1,
		0x02: pixelgl.Key2,
		0x03: pixelgl.Key3,
		0x04: pixelgl.KeyQ,
		0x05: pixelgl.KeyW,
		0x06: pixelgl.KeyE,
		0x07: pixelgl.KeyA,
		0x08: pixelgl.KeyS,
		0x09: pixelgl.KeyD,
		0x0A: pixelgl.KeyZ,
		0x00: pixelgl.KeyX,
		0x0B: pixelgl.KeyC,
		0x0C: pixelgl.Key4,
		0x0D: pixelgl.KeyR,
		0x0E: pixelgl.KeyF,
		0x0F: pixelgl.KeyV,
	}

	return &Window{
		Window:     win,
		KeyMap:     km,
		KeysPushed: [16]*time.Ticker{},
	}, nil
}
