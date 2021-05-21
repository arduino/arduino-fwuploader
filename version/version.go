package version

import "fmt"

var (
	defaultVersionString = "0.0.0-git"
	versionString        = ""
	commit               = ""
	date                 = ""
)

type Info struct {
	Application   string `json:"Application"`
	VersionString string `json:"VersionString"`
	Commit        string `json:"Commit"`
	Date          string `json:"Date"`
}

func NewInfo(application string) *Info {
	return &Info{
		Application:   application,
		VersionString: versionString,
		Commit:        commit,
		Date:          date,
	}
}

func (i *Info) String() string {
	return fmt.Sprintf("%s Version: %s Commit: %s Date: %s", i.Application, i.VersionString, i.Commit, i.Date)
}

func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
}
