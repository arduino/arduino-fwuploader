package arguments

import (
	"github.com/spf13/cobra"
)

// Flags contains various common flags.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Flags struct {
	Address string
	Fqbn    string
}

// AddToCommand adds the flags used to set address and fqbn to the specified Command
func (f *Flags) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:samd:mkr1000, arduino:mbed_nano:nanorp2040connect")
	cmd.Flags().StringVarP(&f.Address, "address", "a", "", "Upload port, e.g.: COM10, /dev/ttyACM0")
}
