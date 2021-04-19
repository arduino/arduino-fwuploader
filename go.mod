module github.com/arduino/FirmwareUpdater

go 1.14

// branch with support for serial timeouts
replace go.bug.st/serial => github.com/cmaglie/go-serial v0.0.0-20200923162623-b214c147e37e

require (
	github.com/arduino/arduino-cli v0.0.0-20210419093035-6ca680d235a3
	github.com/arduino/go-paths-helper v1.4.0
	github.com/imjasonmiller/godice v0.1.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	go.bug.st/serial v1.1.2
)
