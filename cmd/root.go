package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chyp8 [command]",
	Short: "Chip-8 emulator using Go",
	Long:  "A Chip-8 emulator written from scratch that mimics the functionalities of a Chip-8, an interpretted language originally written for the COSMIC-VIP/ Telmac 8 bit systems.",
	Run:   Root,
	Args:  cobra.ExactArgs(1),
}

func Root(cmd *cobra.Command, args []string) {
	fmt.Println("Enter command as `chyp8 start /path/ROM refreshRate`")
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&refreshRate, "refresh", "r", 60, "Set the refresh rate in Hz")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".chyp8" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".chyp8")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
