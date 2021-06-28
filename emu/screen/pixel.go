package screen

import (
	"time"

	"github.com/faiface/pixel/pixelgl"
)

type Window struct {
	*pixelgl.Window
	KeyMap     map[uint16]pixelgl.Button
	KeysPushed [16]*time.Ticker
}
