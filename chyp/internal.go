package chyp

import "os"

type Chyp8 struct {
	opcode      uint16
	memory      [4096]uint8
	V           [16]uint8
	I           uint16
	pc          uint16
	gfx         [64 * 32]uint8
	delay_timer uint8
	sound_timer uint8
	stack       [16]uint16
	sp          uint16
	key         [16]uint8
}

func (emu *Chyp8) Initialize() {
	emu.pc = 0x200
	emu.opcode = 0
	emu.sp = 0
	emu.I = 0
	emu.sound_timer = 0
	emu.delay_timer = 0

	emu.V[0] = 0
	emu.V[1] = 0
	emu.V[2] = 0
	emu.V[3] = 0
	emu.V[4] = 0
	emu.V[5] = 0
	emu.V[6] = 0
	emu.V[7] = 0
	emu.V[8] = 0
	emu.V[9] = 0
	emu.V[10] = 0
	emu.V[11] = 0
	emu.V[12] = 0
	emu.V[13] = 0
	emu.V[14] = 0
	emu.V[15] = 0

}

func LoadGame(filename string) {
	rom, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		panic(filename + "ROM does not exists")
	}
	EmulateCycle(rom)
	defer rom.Close()
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
