# About
Chip-8 is an interpretted language designed to create programs/games on the 8bit systems like the COSMIC VIP.
Chyp8 is an emulator of the Chip-8 system build with Golang.

## How to build it:
Clone the master branch and use ```go run main.go start roms/NAME_OF_THE_ROM -r=REFRESHRATE```
or run the ```.exe``` file in command line as ```./main.exe start roms/NAME_OF_THE_ROM -r=REFRESHRATE```.
The ```-r=REFRESHRATE``` argument is an optional flag which sets the clock cycle of the emulator. 
It's set to ```60``` by default.

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
