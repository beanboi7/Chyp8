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
	V               [16]uint8      //general purpose registers V[x], x = 0-15
	I               uint16         //address register
	pc              uint16         //program counter fetches the next instruction from memory
	display         [64 * 32]uint8 //display represented as an array
	delayTimer      uint8          //counts down at 60Hz
	soundTimer      uint8          //counts down at 60Hz
	stack           [16]uint16     //virtual stack
	sp              uint16         //stack pointer
	keyStates       [16]uint8      //tells whether key is pressed or not
	drawF           bool           //to draw or not to draw
	clock           *time.Ticker   //cpu clock
	audioChannel    chan struct{}
	ShutdownChannel chan struct{}
	window          *screen.Window //gui screen
}

const (
	keyRepeatDuration = time.Second / 1000
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
//load fontset from ROM to memory, returns reference to the emulator, error
func NewEMU(romPath string, clockSpeed int) (*EMU, error) {
	win, err := screen.NewScreen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	emu := EMU{
		memory:          [4096]uint8{},
		V:               [16]uint8{},
		I:               0,
		pc:              0x200,
		display:         [64 * 32]uint8{},
		stack:           [16]uint16{},
		sp:              0,
		keyStates:       [16]uint8{},
		drawF:           false,
		clock:           time.NewTicker(time.Second / time.Duration(clockSpeed)),
		audioChannel:    make(chan struct{}),
		ShutdownChannel: make(chan struct{}),
		window:          win,
	}
	emu.loadFont()

	if err := emu.LoadROM(romPath); err != nil {
		return nil, err
	}

	return &emu, nil

}

//loads fontset to the first 0x200 bytes of memory
func (emu *EMU) loadFont() {
	for i := 0; i < 80; i++ {
		emu.memory[i] = FontSet[i]
	}
}

//reads rom file and loads into memory
func (emu *EMU) LoadROM(path string) error {
	rom, err := ioutil.ReadFile(path)
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
				emu.drawOrUpdate()
				emu.keyPushHandler()

				emu.delayTimerHandler()
				emu.soundTimerHandler()
				continue
			}
			break
		case <-emu.ShutdownChannel:
			break
		}
		break
	}

	emu.shutDownSignal("SayoNara m8")

}

//opcode is loaded from the memory and is sent to the parser
func (emu *EMU) EmulateCycle() {
	emu.opcode = uint16(emu.memory[emu.pc])<<8 | uint16(emu.memory[emu.pc+1])
	emu.drawF = false

	err := emu.opCodeParser()
	if err != nil {
		fmt.Printf("error parsing opcode: %v", err)
	}
}

//when an unidentified opcode is read
func (emu *EMU) opCodeError(opcode uint16) error {
	return fmt.Errorf("Unknown opcode: %x \n", opcode)
}

//closes the audio channel when it reaches zero
func (emu *EMU) soundTimerHandler() {
	if emu.soundTimer > 0 && emu.soundTimer == 1 {
		fmt.Println("BEEP!")
		emu.audioChannel <- struct{}{}
		emu.soundTimer--
	}
}

//delay timer runs until it reaches 0
func (emu *EMU) delayTimerHandler() {
	if emu.delayTimer > 0 {
		emu.delayTimer--
	}
}

//prints the string message and closes the audio channel, sends empty struct to shutdown channel
func (emu *EMU) shutDownSignal(message string) {
	fmt.Println(message)
	close(emu.audioChannel)
	emu.ShutdownChannel <- struct{}{}
}

//supposedly should manage the beep audio in game
//after it decodes and sends it as a stream
func (emu *EMU) ManageAudio() {
	f, err := os.Open("./assets/assets_beep.mp3")
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
	select {}
}

//polls for a kep press on the keyboard
//starts a ticker when a key in the keymap is pressed
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
			// emu.keyStates[key] = 0
		}

	}
}

//if the draw flag is set only then sprites are
//drawn on the gui screen, else update screen
func (emu *EMU) drawOrUpdate() {
	if emu.drawF {
		emu.window.Draw(emu.display)
	} else {
		emu.window.UpdateInput()
	}
}

//in every cycle opcode is sent here to be decoded
//it is then executed accordingly
//need to add comments to each instruction
func (emu *EMU) opCodeParser() error {
	n := emu.opcode & 0x000F
	x := (emu.opcode & 0x0F00) >> 8
	y := (emu.opcode & 0x00F0) >> 4
	kk := emu.opcode & 0x00FF
	F := 0xF

	switch emu.opcode & 0xF000 {
	case 0x0000:
		switch emu.opcode & 0x00FF {
		//clears the screen
		case 0x00E0:
			emu.display = [64 * 32]byte{}
			emu.pc += 2

		//The interpreter sets the program counter to the address at the top of the stack,
		//then subtracts 1 from the stack pointer.
		case 0x00EE:
			emu.pc = emu.stack[emu.sp] + 2
			emu.sp--

		default:
			return emu.opCodeError(emu.opcode & 0x00FF)
		}
	//sets the Program Counter(PC) to nnn
	case 0x1000:
		addr := emu.opcode & 0x0FFF
		emu.pc = addr

	//the stack pointer is incremented, then puts the current PC on the top of the stack.
	//The PC is then set to nnn.
	case 0x2000:
		addr := emu.opcode & 0x0FFF
		emu.sp++
		emu.stack[emu.sp] = emu.pc
		emu.pc = addr

	//Skip next instruction,ie PC+=2 if Vx = kk.
	case 0x3000:
		if emu.V[x] == uint8(kk) {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
	//Skip next instruction if Vx != kk and sets PC+=2
	case 0x4000:
		if emu.V[x] != uint8(kk) {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
	//Skip next instruction if Vx = Vy and sets PC+=2
	case 0x5000:
		if emu.V[x] == emu.V[y] {
			emu.pc += 4
		} else {
			emu.pc += 2
		}
	//6xkk - LD Vx, byte
	//Sets Vx = kk.
	case 0x6000:
		emu.V[x] = uint8(kk)
		emu.pc += 2

	//7xkk - ADD Vx, byte
	//Set Vx = Vx + kk.
	case 0x7000:
		emu.V[x] += uint8(kk)
		emu.pc += 2

	case 0x8000:
		switch n {
		//Stores the value of register Vy in register Vx.
		case 0x0000:
			emu.V[x] = emu.V[y]
			emu.pc += 2

		//Set Vx = Vx OR Vy; OR = bitwise OR operation
		case 0x0001:
			emu.V[x] |= emu.V[y]
			emu.pc += 2

		//Set Vx = Vx AND Vy; AND = bitwise AND operation
		case 0x0002:
			emu.V[x] &= emu.V[y]
			emu.pc += 2

		//Set Vx = Vx XOR Vy; XOR = bitwise XOR operation
		case 0x0003:
			emu.V[x] ^= emu.V[y]
			emu.pc += 2

		//Set Vx = Vx + Vy, set VF = carry.
		case 0x0004:
			if emu.V[x]+emu.V[y] > 0x00FF {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.V[x] += emu.V[y]
			emu.pc += 2

		//If Vx > Vy, then VF is set to 1, otherwise 0.
		//Then Vy is subtracted from Vx, and the results stored in Vx.
		case 0x0005:
			if emu.V[x] > emu.V[y] {
				emu.V[F] = 1
			} else {
				emu.V[F] = 0
			}
			emu.V[x] -= emu.V[y]
			emu.pc += 2

		//If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
		//Then Vx is divided by 2.
		case 0x0006:
			emu.V[x] = emu.V[y] >> 1
			emu.V[F] = emu.V[y] & 0x01
			emu.pc += 2

		//If Vy > Vx, then VF is set to 1, otherwise 0.
		//Then Vx is subtracted from Vy, and the results stored in Vx.
		case 0x0007:
			emu.V[x] = emu.V[y] - emu.V[x]
			emu.pc += 2

		//If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
		//Then Vx is multiplied by 2.
		case 0x000E:
			emu.V[x] = emu.V[y] << 1   //multiplying with two
			emu.V[F] = emu.V[y] & 0x80 //gets MSB
			emu.pc += 2

		default:
			return emu.opCodeError(emu.opcode & 0x000F)
		}

	//Skip next instruction if Vx != Vy.
	case 0x9000:
		if emu.V[x] != emu.V[y] {
			emu.pc += 4
		} else {
			emu.pc += 2
		}

	//The value of register I is set to nnn.
	case 0xA000:
		addr := emu.opcode & 0x0FFF
		emu.I = addr
		emu.pc += 2

	//The program counter is set to nnn plus the value of V0.
	case 0xB000:
		addr := emu.opcode & 0x0FFF
		emu.pc = uint16(emu.V[0]) + addr
		emu.pc += 2

	//A random number is generated from 0 to 255,
	//which is then AND-ed with the value kk. The results are stored in Vx.
	case 0xC000:
		emu.V[x] = uint8(uint16(rand.Intn(255-0)+0) & kk)
		emu.pc += 2

	//Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
	case 0xD000:
		//draw sprites on the screen
		var rows uint16
		var columns uint16
		emu.V[F] = 0
		for rows = 0; rows < n; rows++ {
			spriteData := uint16(emu.memory[emu.I+rows]) //sprite data is stored in the mem[I+yline] address
			for columns = 0; columns < 8; columns++ {
				pixel := uint16(emu.V[x]) + columns + ((uint16(emu.V[y]) + rows) * 64)

				if pixel >= uint16(len(emu.display)) {
					break
				}
				//To check if the read data from mem[I + ...] and the gfx[] data are set to 1, if yes then we need its 1XOR1 ,ie collision
				if spriteData&(0x80>>columns) != 0 { //if the bit value read from mem is not 0 and
					if emu.display[pixel] == 1 { // bitwise value already on the screen is set to 1, then collison is detected and thus 1^1, ie XOR-ed and the V[F] = 1
						emu.V[F] = 1 // collison detected
					}
					emu.display[pixel] ^= 1
				}

			}
		}
		emu.drawF = true
		emu.pc += 2

	case 0xE000:
		switch kk {
		// Skip next instruction if key with the value of Vx is pressed.
		case 0x009E:
			if emu.keyStates[emu.V[x]] == 1 {
				emu.pc += 4
				emu.keyStates[emu.V[x]] = 0
			} else {
				emu.pc += 2
			}

		//Skip next instruction if key with the value of Vx is not pressed.
		case 0x00A1:
			if emu.keyStates[emu.V[x]] == 0 {
				emu.pc += 4
			} else {
				emu.keyStates[emu.V[x]] = 0
				emu.pc += 2
			}

		default:
			return emu.opCodeError(emu.opcode)
		}

	case 0xF000:
		switch kk {
		//Set Vx = delay timer value.
		case 0x0007:
			emu.V[x] = emu.delayTimer
			emu.pc += 2

		//Awaits a keypress by checking value in the array, if true then sets to Vx
		case 0x000A:
			for i, val := range emu.keyStates {
				if val != 0 {
					emu.V[x] = uint8(i)
					emu.pc += 2
				}
			}
			emu.keyStates[emu.V[x]] = 0

		//Set delay timer = Vx.
		case 0x0015:
			emu.delayTimer = emu.V[x]
			emu.pc += 2

		//Set sound timer = Vx.
		case 0x0018:
			emu.soundTimer = emu.V[x]
			emu.pc += 2

		//Set I = I + Vx.
		case 0x001E:
			emu.I += uint16(emu.V[x])
			emu.pc += 2

		//Sets I to the location of the sprite for the character in VX.
		//Characters 0-F (in hexadecimal) are represented by a 4x5 font
		case 0x0029:
			emu.I = uint16(emu.V[x]) * 0x5
			emu.pc += 2

		//Store BCD representation of Vx in memory locations I, I+1, and I+2.
		case 0x0033:
			emu.memory[emu.I] = emu.V[x] / 100
			emu.memory[emu.I+1] = (emu.V[x] / 10) % 10
			emu.memory[emu.I+2] = (emu.V[x] % 100) % 10
			emu.pc += 2

		//Store registers V0 through Vx in memory starting at location I.
		case 0x0055:
			for i := uint16(0); i <= x; i++ {
				emu.memory[emu.I+i] = emu.V[i]
			}
			emu.pc += 2

		//Read registers V0 through Vx from memory starting at location I.
		case 0x0065:
			for i := uint16(0); i <= x; i++ {
				emu.V[i] = emu.memory[emu.I+i]
			}
			emu.pc += 2

		default:
			return emu.opCodeError(emu.opcode & 0x00FF)
		}
	default:
		return emu.opCodeError(emu.opcode)
	}
	return nil
}
