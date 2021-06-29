package screen

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Window struct {
	*pixelgl.Window
	KeyMap     map[uint16]pixelgl.Button
	KeysPushed [16]*time.Ticker
	Key        pixelgl.Button
}

func Screen() {
	cfg := pixelgl.WindowConfig{
		Title:                  "Chyp8",
		Icon:                   []pixel.Picture{},
		Bounds:                 pixel.R(0, 0, 512, 346),
		Position:               pixel.Vec{},
		Monitor:                &pixelgl.Monitor{},
		Resizable:              false,
		Undecorated:            false,
		NoIconify:              false,
		AlwaysOnTop:            false,
		TransparentFramebuffer: false,
		VSync:                  true,
		Maximized:              false,
		Invisible:              false,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		win.Update()
	}
}
