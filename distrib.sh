#!/bin/bash

rm -rf distrib
mkdir -p distrib/linux64
mkdir -p distrib/linux32
mkdir -p distrib/linuxarm
mkdir -p distrib/linuxarm64
mkdir -p distrib/osx
mkdir -p distrib/windows

export CGO_ENABLED=0

GOOS=linux GOARCH=amd64 go build -o distrib/linux64/updater github.com/arduino-libraries/FirmwareUpdater/cli
GOOS=linux GOARCH=386 GO386=387 go build -o distrib/linux32/updater github.com/arduino-libraries/FirmwareUpdater/cli
GOOS=linux GOARCH=arm go build -o distrib/linuxarm/updater github.com/arduino-libraries/FirmwareUpdater/cli
GOOS=linux GOARCH=arm64 go build -o distrib/linuxarm64/updater github.com/arduino-libraries/FirmwareUpdater/cli
GOOS=windows GOARCH=386 GO386=387 go build -o distrib/windows/updater.exe github.com/arduino-libraries/FirmwareUpdater/cli

#export CGO_ENABLED=1
# need osxcross in path
GOOS=darwin GOARCH=amd64 go build -o distrib/osx/updater github.com/arduino-libraries/FirmwareUpdater/cli

cp -r $GOPATH/src/github.com/arduino-libraries/FirmwareUpdater/firmwares distrib/

# call the tool with something like
# ./linux64/updater -flasher firmwares/NINA/FirmwareUpdater.mkrwifi1010.ino.bin -firmware firmwares/NINA/1.2.1/NINA_W102.bin -port /dev/ttyACM0  -address arduino.cc:443 -restore_binary /tmp/arduino_build_619137/WiFiSSLClient.ino.bin -programmer {runtime.tools.bossac}/bossac