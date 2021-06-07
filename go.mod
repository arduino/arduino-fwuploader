module github.com/arduino/FirmwareUploader

go 1.14

// branch with support for serial timeouts
replace go.bug.st/serial => github.com/cmaglie/go-serial v0.0.0-20200923162623-b214c147e37e

require (
	github.com/arduino/arduino-cli v0.0.0-20210422154105-5aa424818026
	github.com/arduino/go-paths-helper v1.4.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.6.1
	go.bug.st/serial v1.1.2
)
