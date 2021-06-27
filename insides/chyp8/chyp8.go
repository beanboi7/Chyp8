package chyp

import (
	"fmt"
	"io/ioutil"
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
	updateScreen    bool
	audioChannel    chan struct{}
	shutdownChannel chan struct{}
	window          *screen.Window
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
		updateScreen:    false,
		audioChannel:    make(chan struct{}),
		shutdownChannel: make(chan struct{}),
		window:          &screen.Window{},
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

//after creating a new emu, it should be Run

func (emu *EMU) Run() {
	for {
		emu.EmulateCycle()
	}

}

func (emu *EMU) EmulateCycle() {
	emu.opcode = uint16(emu.memory[emu.pc]<<8 | emu.memory[emu.pc+1])
	emu.updateScreen = false

	err := emu.opCodeParser()
	if err != nil {
		fmt.Printf("error parsing opcode: %v", err)
	}
}

func (emu *EMU) opCodeParser() error {
	//big shit logic for parsing the opcode
}
