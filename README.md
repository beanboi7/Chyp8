[![GitHub license](https://img.shields.io/github/license/beanboi7/Chyp8?color=pink&logo=github)](https://github.com/beanboi7/Chyp8/blob/master/LICENSE) [![Releases](https://img.shields.io/badge/status-building-brightgreen)](https://github.com/beanboi7/Chyp8/releases)
# Introduction
Chip-8 is an interpretted language designed to create programs/games on the 8bit systems like the COSMAC VIP and Telmac 1800.
Chyp8 is an emulator of the Chip-8 system build with Golang. For more info on Chip-8 refer [here](https://en.wikipedia.org/wiki/CHIP-8)

## To build Chyp8:
Clone the master branch and use ```go run main.go start roms/NAME_OF_THE_ROM -r=REFRESHRATE```
or run the ```.exe``` file in command line as ```./main.exe start roms/NAME_OF_THE_ROM -r=REFRESHRATE```.
The ```-r=REFRESHRATE``` argument is an optional flag which sets the clock cycle of the emulator. 
It's set to ```60``` by default.

## To Use Chyp8 (without building it):
Get the ```tar.gz``` or ```zip``` folder downloaded from the ```Releases``` according to your OS and run ```./main.exe start roms/NAME_OF_THE_ROM -r=REFRESHRATE``` in a terminal.

## Screenshots:
- ### Opcode testing
    ![Opcodes tested image](./assets/opcodetest.png)
- ### Tetris
    ![tetris](./assets/tetris.png)

## KeyMap

<pre>  <b>Keypad</b>                 <b>Keyboard</b> 
             
|1|2|3|C|                |1|2|3|4|
|4|5|6|D|                |Q|W|E|R|
                =>
|7|8|9|E|                |A|S|D|F|
|A|0|B|F|                |Z|X|C|V|</pre>


## Packages used:
- PixelGL
- Cobra

## Issues
- [ ] Fix audio bug
- [ ] Find correct ticker for the keypad

## References 
- [Laurence Muller's post on writing an emulator](https://multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter/) 
- [Cowgod's chip-8 technical reference manual](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#2.4)
- [Chippy](https://github.com/bradford-hamilton/chippy)
