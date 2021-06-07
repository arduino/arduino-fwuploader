module github.com/arduino/FirmwareUploader

go 1.14

// branch with support for serial timeouts
replace go.bug.st/serial => github.com/cmaglie/go-serial v0.0.0-20200923162623-b214c147e37e

require (
	github.com/arduino/arduino-cli v0.0.0-20210603144340-aef5a54882fa
	github.com/arduino/go-paths-helper v1.6.0
	github.com/cmaglie/go.rice v1.0.3
	github.com/mattn/go-colorable v0.1.8
	github.com/pkg/errors v0.9.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.6.1
	go.bug.st/downloader/v2 v2.1.1
	go.bug.st/relaxed-semver v0.0.0-20190922224835-391e10178d18
	go.bug.st/serial v1.1.2
)
