package cmd

import (
	"fmt"
	"os"

	"chyp8/emu/cpu"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start `path/ROM`",
	Short: "load and start the Emulator",
	Args:  cobra.MinimumNArgs(1),
	Run:   Start,
}

// chyp8 start 'path/to/ROM' -r 69
var refreshRate int

func init() {
	rootCmd.AddCommand(startCmd)
}

func Start(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Enter the path to the ROM as: `path/ROM` ")
		os.Exit(1)
	}

	romPath := os.Args[2]

	emu, err := cpu.NewEMU(romPath, refreshRate)
	if err != nil {
		fmt.Printf("\n Error starting the Emulator:%v \n", err)
	}
	go emu.Run()
	go emu.ManageAudio()

	<-emu.ShutdownChannel
}
