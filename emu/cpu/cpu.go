package chyp

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"chyp8/emu/screen"

	"github.com/faiface/pixel/pixelgl"
)

type EMU struct {
	opcode          uint16
	memory          [4096]uint8
	V               [16]uint8
	I               uint16 //address register
	pc              uint16
	display         [64 * 32]uint8
	delayTimer      uint8 //counts down at 60Hz
	soundTimer      uint8 //same as above
	stack           [16]uint16
	sp              uint16
	keyState        [16]uint8 //tells whether key is pressed or not
	updateScreen    bool      //to draw or not
	audioChannel    chan struct{}
	shutdownChannel chan struct{}
	window          *screen.Window
}

const (
	keyRepeatDuration = time.Second / 5
	maxRomSize        = 0xFFF - 0x200
)

var FontSet = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func (emu *EMU) keyPress() uint16 {
	emu.window.KeyMap[uint16(pixelgl.Key1)] = 0x01
	emu.window.KeyMap[uint16(pixelgl.Key2)] = 0x02
	emu.window.KeyMap[uint16(pixelgl.Key3)] = 0x03
	emu.window.KeyMap[uint16(pixelgl.KeyQ)] = 0x04
	emu.window.KeyMap[uint16(pixelgl.KeyW)] = 0x05
	emu.window.KeyMap[uint16(pixelgl.KeyE)] = 0x06
	emu.window.KeyMap[uint16(pixelgl.KeyA)] = 0x07
	emu.window.KeyMap[uint16(pixelgl.KeyS)] = 0x08
	emu.window.KeyMap[uint16(pixelgl.KeyD)] = 0x09
	emu.window.KeyMap[uint16(pixelgl.KeyZ)] = 0x0A
	emu.window.KeyMap[uint16(pixelgl.KeyX)] = 0x00
	emu.window.KeyMap[uint16(pixelgl.KeyC)] = 0x0B
	emu.window.KeyMap[uint16(pixelgl.Key4)] = 0x0C
	emu.window.KeyMap[uint16(pixelgl.KeyR)] = 0x0D
	emu.window.KeyMap[uint16(pixelgl.KeyF)] = 0x0E
	emu.window.KeyMap[uint16(pixelgl.KeyV)] = 0x0F

	switch pixelgl.JustPressed(pixelgl.Button) {
	case pixelgl.Key1:
		return emu.window.KeyMap[uint16(pixelgl.Key1)]
		break
	case pixelgl.Key2:
		return
	}
}

//should init window and EMU,
//load fontset from ROM to mem, return pointer to VM or an error
func NewEMU(romPath string, clockSpeed int) (*EMU, error) {

	emu := EMU{
		opcode:          0,
		memory:          [4096]uint8{},
		V:               [16]uint8{},
		I:               0,
		pc:              0x200,
		display:         [2048]uint8{},
		delayTimer:      0,
		soundTimer:      0,
		stack:           [16]uint16{},
		sp:              0,
		keyState:        [16]uint8{},
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
		emu.memory[i] = FontSet[i]
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

	return nil
}

func (emu *EMU) Run() {
	for {

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
	n := emu.opcode & 0x000F
	x := emu.opcode & 0x0F00
	y := emu.opcode & 0x00F0
	kk := emu.opcode & 0x00FF
	F := 15

	switch emu.opcode & 0x0FFF {
	case 0x00E0:
		emu.gfx = [64 * 32]byte{}
		emu.pc += 2
		break
	case 0x00EE:
		emu.pc = emu.stack[emu.sp]
		emu.sp--
		break
	}

	switch emu.opcode & 0xF000 {
	case 0x1000:
		addr := emu.opcode & 0x0FFF
		emu.pc = addr
		break
	case 0x2000:
		addr := emu.opcode & 0x0FFF
		emu.pc = addr
		emu.sp++
		emu.stack[emu.sp] = emu.pc
		break
	case 0x3000:
		if emu.V[x] == uint8(kk) {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
		break
	case 0x4000:
		if emu.V[x] != uint8(kk) {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
		break
	case 0x5000:
		if emu.V[x] == emu.V[y] {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
		break
	case 0x6000:
		emu.V[x] = uint8(kk)
		emu.pc += 2
		break
	case 0x7000:
		emu.V[x] += uint8(kk)
		emu.pc += 2
		break
	case 0x8000:
		switch n {
		case 0:
			emu.V[x] = emu.V[y]
			emu.pc += 2
			break
		case 1:
			emu.V[x] |= emu.V[y]
			emu.pc += 2
			break
		case 2:
			emu.V[x] &= emu.V[y]
			emu.pc += 2
			break
		case 3:
			emu.V[x] ^= emu.V[y]
			emu.pc += 2
			break
		case 4:
			emu.V[x] += emu.V[y]
			if emu.V[x] >= 0xFF {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.pc += 2
			break
		case 5:
			if emu.V[x] > emu.V[y] {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.V[x] -= emu.V[y]
			emu.pc += 2
			break
		case 6:
			//Vx>>=1
			if emu.V[x]&00000001 == 1 {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.V[x] /= 2
			emu.pc += 2
			break
		case 7:
			if emu.V[y] > emu.V[x] {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.V[x] = emu.V[y] - emu.V[x]
			emu.pc += 2
			break

		case 0x000E:
			//Vx<<=1
			if emu.V[x]&64 == 1 {
				emu.V[F] = 1
				emu.V[x] *= 2
			} else {
				emu.V[F] = 0
			}
			emu.pc += 2
			break
		}
		break
	case 0x9000:
		if emu.V[x] != emu.V[y] {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
		break
	case 0xA000:
		addr := emu.opcode & 0x0FFF
		emu.I = addr
		emu.pc += 2
		break
	case 0xB000:
		addr := emu.opcode & 0x0FFF
		emu.pc = uint16(emu.V[0] + uint8(addr))
		emu.pc += 2
		break
	case 0xC000:
		random := rand.Intn(255-0) + 0
		emu.V[x] = uint8(random & int(kk))
		emu.pc += 2
		break
	case 0xD000:
		//draw sprites on the screen
	case 0xE000:
		//make the keyboard mapping and come here
		switch kk {
		case 0x9E:

		}

	case 0xF000:
		addr := emu.opcode & 0x00FF
		switch addr {
		case 0x07:
			// delay_timer shit,need more clarity on how to setup DT
			emu.V[x] = emu.delayTimer
			emu.pc += 2
			break
		case 0x0A:
			//keypress clarity needed
		case 0x15:
			emu.delayTimer = emu.V[x]
			emu.pc += 2
			break
		case 0x18:
			emu.soundTimer = emu.V[x]
			emu.pc += 2
			break
		case 0x1E:
			emu.I = emu.I + uint16(emu.V[x])
			emu.pc += 2
			break
		case 0x29:
			//sprites and shit
		case 0x33:
			// emu.memory[emu.I] =
			// emu.memory[emu.I + 1] = emu.V[x]

		case 0x55:
			for i := 0; i < int(x); i++ {
				emu.memory[emu.I+uint16(i)] = emu.V[i]
			}
			emu.pc += 2
			break
		case 0x65:
			for i := 0; i < int(x); i++ {
				emu.V[i] = emu.memory[emu.I+uint16(i)]
			}
			emu.pc += 2
			break
		}
	}
	return nil
}