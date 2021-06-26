package chyp

import (
	"chyp-8/insides/screen"
	"io/ioutil"
	"os"
	"time"

	"github.com/beanboi7/chyp-8/insides/screen"
)

type EMU struct {
	opcode          uint16
	memory          [4096]uint8
	V               [16]uint8
	I               uint16
	pc              uint16
	gfx             [64 * 32]uint8
	delay_timer     uint8
	sound_timer     uint8
	stack           [16]uint16
	sp              uint16
	keyboard        [16]uint8
	draw            bool
	audioChannel    chan struct{}
	shutdownChannel chan struct{}
}

const (
	keyRepeatDuration = time.Second / 5
	maxRomSize        = 0xFFF - 0x200
)

//should init window and EMU,
//load fontset from ROM to mem, return pointer to VM or an error
func NewEMU(romPath string, clockSpeed int) (*EMU, error) {

	emu := EMU{
		opcode:          0,
		memory:          [4096]uint8{},
		V:               [16]uint8{},
		I:               0,
		pc:              0x200,
		gfx:             [2048]uint8{},
		delay_timer:     0,
		sound_timer:     0,
		stack:           [16]uint16{},
		sp:              0,
		keyboard:        [16]uint8{},
		draw:            false,
		audioChannel:    make(chan struct{}),
		shutdownChannel: make(chan struct{}),
	}

	emu.loadFont()

	err := emu.LoadROM(romPath)
	if err != nil {
		return nil, err
	}
	return &emu, nil

}

func (emu *EMU) loadFont() {
	for i := 0; i < 80; i++ {
		emu.memory[i] = screen.FontSet[i]
	}
}

func (emu *EMU) LoadROM(filename string) error {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if len(rom) > maxRomSize {
		panic("ROM too big, can't cross 3583 bytes")
	}

	//otherwise its loaded properly, load rom to memory
	for i := 0; i < len(rom); i++ {
		emu.memory[0x200+i] = rom[i]
	}
}

func EmulateCycle(file *os.File) {

	//Fetch
	//Decode
	//Execute

	//load fontset from memory

	//Reset timers

}

func SetKeys() {

}
