module github.com/arduino/FirmwareUpdater

go 1.14

// branch with support for serial timeouts
replace go.bug.st/serial => github.com/cmaglie/go-serial v0.0.0-20200923162623-b214c147e37e

require (
	github.com/arduino/arduino-cli v0.0.0-20200924151007-69ac12c98b2b
	github.com/pkg/errors v0.9.1
	go.bug.st/serial v1.1.1
)
