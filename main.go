package main

import (
	"os"

	"github.com/arduino/FirmwareUploader/cli"
)

func main() {
	uploaderCmd := cli.NewCommand()
	if err := uploaderCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
