package version

import "fmt"

var (
	defaultVersionString = "0.0.0-git"
	versionString        = ""
	commit               = ""
	date                 = ""
)

func String() string {
	return fmt.Sprintf("FirmwareUploader Version: %s Commit: %s Date: %s", versionString, commit, date)
}

func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
}
