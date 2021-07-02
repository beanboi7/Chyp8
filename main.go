package main

import (
	"chyp8/cmd"

	"github.com/faiface/pixel/pixelgl"
)

func main() {
	pixelgl.Run(runChyp8)
}

func runChyp8() {
	cmd.Execute()
}
