package version

import (
	"os"

	v "github.com/arduino/FirmwareUploader/version"
	"github.com/arduino/arduino-cli/cli/feedback"
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

func run(cmd *cobra.Command, args []string) {
	feedback.Print(v.VersionInfo)
}
