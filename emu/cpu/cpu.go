package cpu

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"chyp8/emu/screen"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
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
	keyStates       [16]uint8 //tells whether key is pressed or not
	drawF           bool      //to draw or not
	clock           *time.Ticker
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
		keyStates:       [16]uint8{},
		drawF:           false,
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
		select {
		case <-emu.clock.C:
			if !emu.window.Closed() {
				emu.EmulateCycle()
				emu.keyPushHandler()
				emu.drawOrUpdate()
				emu.soundTimerHandler()
				emu.delayTimerHandler()
				continue
			}
			break
		case <-emu.shutdownChannel:
			break
		}
		break
	}

	emu.shutDownSignal("bye")

}

func (emu *EMU) EmulateCycle() {
	emu.opcode = uint16(emu.memory[emu.pc]<<8 | emu.memory[emu.pc+1])
	emu.drawF = false

	err := emu.opCodeParser()
	if err != nil {
		fmt.Printf("error parsing opcode: %v", err)
	}
}

func (emu *EMU) opCodeParser() error {
	n := emu.opcode & 0x000F
	x := (emu.opcode & 0x0F00) >> 8 // right shift because V register has index from 0-15 only ad
	y := (emu.opcode & 0x00F0) >> 4 // similar right shift to get the value of y to lie b/w 0 - 15
	kk := emu.opcode & 0x00FF
	F := 15

	switch emu.opcode & 0x0FFF {
	case 0x00E0:
		emu.display = [64 * 32]byte{}
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
		case 0x0000:
			emu.V[x] = emu.V[y]
			emu.pc += 2
			break
		case 0x0001:
			emu.V[x] |= emu.V[y]
			emu.pc += 2
			break
		case 0x0002:
			emu.V[x] &= emu.V[y]
			emu.pc += 2
			break
		case 0x0003:
			emu.V[x] ^= emu.V[y]
			emu.pc += 2
			break
		case 0x0004:
			emu.V[x] += emu.V[y]
			if emu.V[x] >= 0x00FF {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.pc += 2
			break
		case 0x0005:
			if emu.V[x] > emu.V[y] {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.V[x] -= emu.V[y]
			emu.pc += 2
			break
		case 0x0006:
			//Vx>>=1
			if emu.V[x]&0x1 == 1 {
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
			//basically taking the value of MSB and storing in VF register
			emu.V[F] = emu.V[x] >> 7
			emu.V[x] <<= 1 //multiplying with two
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
		emu.drawF = true //need to draw for this cycle
		var rows uint16
		var columns uint8
		emu.V[F] = 0
		for rows = 0; rows < n; rows++ {
			spriteData := emu.memory[emu.I+rows] //sprite data is stored in the mem[I+yline] address
			for columns = 0; columns < 8; columns++ {
				//now to check if the read data from mem[I + ...] and the gfx[] data are set to 1, if yes then we need its 1XOR1 ,ie collision
				if spriteData&(0x80>>columns) != 0 { //if the bit value read from mem is not 0 and
					if emu.display[x+uint16(columns)+64*(y+rows)] == 1 { // bitwise value already on the screen is set to 1, then collison is detected and thus 1^1, ie XOR-ed and the V[F] = 1
						emu.V[F] = 1 // collison detected
					}
					emu.display[x+uint16(columns)+64*(y+rows)] ^= 1
				}
			}
		}
		emu.pc += 2
	case 0xE000:
		//make the keyboard mapping and come here
		switch kk {
		case 0x9E:
			if emu.keyStates[emu.V[x]] != 0 {
				emu.pc += 4
			} else {
				emu.pc += 2
			}
			break
		case 0xA1:
			if emu.keyStates[emu.V[x]] == 0 {
				emu.pc += 4
			} else {
				emu.pc += 2
			}
			break
		}

	case 0xF000:
		addr := emu.opcode & 0x00FF
		switch addr {
		case 0x07:
			emu.V[x] = emu.delayTimer
			emu.pc += 2
			break
		case 0x0A:
			//awaits a keypress by checking value in the array, if true then sets the index to Vx
			for i, val := range emu.keyStates {
				if val != 0 {
					emu.V[x] = uint8(i)
				}
			}
			emu.pc += 2
			break
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
			if emu.I > 0xFFF {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.pc += 2
			break
		case 0x29:
			//no idea what this instruction does
			emu.I = uint16(emu.V[x] * 0x5)
			emu.pc += 2
		case 0x33:
			emu.memory[emu.I] = emu.V[x] / 100
			emu.memory[emu.I+1] = (emu.V[x] / 10) % 10
			emu.memory[emu.I+2] = (emu.V[x] % 100) % 10
			emu.pc += 2
			break

		case 0x55:
			for i := 0; i < int(x+1); i++ {
				emu.memory[emu.I+uint16(i)] = emu.V[i]
			}
			emu.pc += 2
			break
		case 0x65:
			for i := 0; i < int(x+1); i++ {
				emu.V[i] = emu.memory[emu.I+uint16(i)]
			}
			emu.pc += 2
			break
		}
	default:
		return emu.opCodeError(emu.opcode)
	}
	return nil
}

func (emu *EMU) opCodeError(opcode uint16) error {
	return fmt.Errorf("Unknown opcode: %x", opcode)
}

//closes the audio channel when it reaches zero
func (emu *EMU) soundTimerHandler() {
	if emu.soundTimer > 0 && emu.soundTimer == 1 {
		fmt.Println("BEEP!")
		emu.audioChannel <- struct{}{}
		emu.soundTimer--
	}
}

func (emu *EMU) delayTimerHandler() {
	if emu.delayTimer > 0 {
		emu.delayTimer--
	}
}

//prints the string message and closes the audio channel, sends empty struct to shutdown channel
func (emu *EMU) shutDownSignal(message string) {
	fmt.Println(message)
	close(emu.audioChannel)
	emu.shutdownChannel <- struct{}{}
}

func (emu *EMU) ManageAudio() {
	f, err := os.Open("chyp8/assets/assets_beep.mp3")
	if err != nil {
		fmt.Print(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		fmt.Print(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(streamer)
}

func (emu *EMU) keyPushHandler() {
	for key, val := range emu.window.KeyMap { //iterating over the keymap to check if something is pressed down
		if emu.window.JustReleased(val) && emu.window.KeysPushed[key] != nil { //if a key is released, ie no longer pressed, then stop the ticker and assign it to nil
			emu.window.KeysPushed[key].Stop()
			emu.window.KeysPushed[key] = nil
		} else if emu.window.JustPressed(val) && emu.window.KeysPushed[key] == nil { //else if a key is pressed and its ticker is zero, start a new ticker for that corresponding key
			emu.window.KeysPushed[key] = time.NewTicker(keyRepeatDuration)
			emu.keyStates[key] = 1
		}

		if emu.window.KeysPushed[key] == nil {
			continue //when the ticker is at nil and none of the keys are pressed
		}

		select {
		case <-emu.window.KeysPushed[key].C:
			emu.keyStates[key] = 1
		default:
			emu.keyStates[key] = 0
		}

	}
}

func (emu *EMU) drawOrUpdate() {
	if emu.drawF {
		emu.window.Draw(emu.display)
	} else {
		emu.window.Update()
	}
}
