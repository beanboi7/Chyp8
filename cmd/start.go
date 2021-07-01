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

	go emu.ManageAudio()
	go emu.Run()
}

var refreshRate int

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&refreshRate, "refresh", "r", 60, "sets the refresh rate of the display")
}
