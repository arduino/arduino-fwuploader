package version

import (
	"fmt"
	"os"

	v "github.com/arduino/FirmwareUploader/version"
	"github.com/spf13/cobra"
)

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Shows version number of FirmwareUploader.",
		Long:    "Shows the version number of FirmwareUploader which is installed on your system.",
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run:     run,
	}
}

var VersionInfo = v.NewInfo("FirmwareUploader")

func run(cmd *cobra.Command, args []string) {
	fmt.Print(VersionInfo)
}
