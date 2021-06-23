package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

//dummy file, will be using pixelgl for UI

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Vanga nanba!",
		Bounds: pixel.R(0, 0, 512, 384),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	for !win.Closed() {
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
