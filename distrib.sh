#!/bin/bash -ex

VERSION=`git describe --tags`
FILENAME="FirmwareUploader"

rm -rf distrib
mkdir -p distrib/linux64
mkdir -p distrib/linux32
mkdir -p distrib/linuxarm
mkdir -p distrib/linuxarm64
mkdir -p distrib/osx
mkdir -p distrib/windows32
mkdir -p distrib/windows64

export CGO_ENABLED=0

GOOS=linux GOARCH=amd64 go build -o distrib/linux64/${FILENAME}
GOOS=linux GOARCH=386 GO386=387 go build -o distrib/linux32/${FILENAME}
GOOS=linux GOARCH=arm go build -o distrib/linuxarm/${FILENAME}
GOOS=linux GOARCH=arm64 go build -o distrib/linuxarm64/${FILENAME}
GOOS=windows GOARCH=386 GO386=387 go build -o distrib/windows32/${FILENAME}.exe
GOOS=windows GOARCH=amd64 go build -o distrib/windows64/${FILENAME}.exe

#export CGO_ENABLED=1
# need osxcross in path
GOOS=darwin GOARCH=amd64 go build -o distrib/osx/${FILENAME}

cp -r firmwares distrib/linux64
cp -r firmwares distrib/linux32
cp -r firmwares distrib/linuxarm
cp -r firmwares distrib/linuxarm64
cp -r firmwares distrib/windows32
cp -r firmwares distrib/windows64
cp -r firmwares distrib/osx

cd distrib/linux64 && tar cjf ../${FILENAME}_${VERSION}_Linux_64bit.tar.bz2 * && cd -
LINUX64_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_Linux_64bit.tar.bz2 | cut -f1 -d " "`
LINUX64_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_Linux_64bit.tar.bz2 | cut -f5 -d " "`

cd distrib/linux32 && tar cjf ../${FILENAME}_${VERSION}_Linux_32bit.tar.bz2 * && cd -
LINUX32_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_Linux_32bit.tar.bz2 | cut -f1 -d " "`
LINUX32_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_Linux_32bit.tar.bz2 | cut -f5 -d " "`

cd distrib/linuxarm && tar cjf ../${FILENAME}_${VERSION}_Linux_ARM.tar.bz2 * && cd -
LINUXARM_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_Linux_ARM.tar.bz2 | cut -f1 -d " "`
LINUXARM_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_Linux_ARM.tar.bz2 | cut -f5 -d " "`

cd distrib/linuxarm64 && tar cjf ../${FILENAME}_${VERSION}_Linux_ARM64.tar.bz2 * && cd -
LINUXARM64_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_Linux_ARM64.tar.bz2 | cut -f1 -d " "`
LINUXARM64_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_Linux_ARM64.tar.bz2 | cut -f5 -d " "`

cd distrib/osx && tar cjf ../${FILENAME}_${VERSION}_macOS_64bit.tar.bz2 * && cd -
OSX_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_macOS_64bit.tar.bz2 | cut -f1 -d " "`
OSX_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_macOS_64bit.tar.bz2 | cut -f5 -d " "`

cd distrib/windows32 && zip -r ../${FILENAME}_${VERSION}_Windows_32bit.zip * && cd -
WINDOWS32_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_Windows_32bit.zip | cut -f1 -d " "`
WINDOWS32_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_Windows_32bit.zip | cut -f5 -d " "`

cd distrib/windows64 && zip -r ../${FILENAME}_${VERSION}_Windows_64bit.zip * && cd -
WINDOWS64_SHA=`sha256sum distrib/${FILENAME}_${VERSION}_Windows_64bit.zip | cut -f1 -d " "`
WINDOWS64_SIZE=`ls -la distrib/${FILENAME}_${VERSION}_Windows_64bit.zip | cut -f5 -d " "`


echo "=============================="
echo "BOARD MANAGER SNIPPET"
echo "=============================="

cat extras/package_index.json.template |
sed "s/%%VERSION%%/${VERSION}/" |
sed "s/%%FILENAME%%/${FILENAME}/" |
sed "s/%%LINUX64_SHA%%/${LINUX64_SHA}/" |
sed "s/%%LINUX64_SIZE%%/${LINUX64_SIZE}/" |
sed "s/%%LINUX32_SHA%%/${LINUX32_SHA}/" |
sed "s/%%LINUX32_SIZE%%/${LINUX32_SIZE}/" |
sed "s/%%LINUXARM_SHA%%/${LINUXARM_SHA}/" |
sed "s/%%LINUXARM_SIZE%%/${LINUXARM_SIZE}/" |
sed "s/%%LINUXARM64_SHA%%/${LINUXARM64_SHA}/" |
sed "s/%%LINUXARM64_SIZE%%/${LINUXARM64_SIZE}/" |
sed "s/%%OSX_SHA%%/${OSX_SHA}/" |
sed "s/%%OSX_SIZE%%/${OSX_SIZE}/" |
sed "s/%%WINDOWS32_SHA%%/${WINDOWS32_SHA}/" |
sed "s/%%WINDOWS32_SIZE%%/${WINDOWS32_SIZE}/" |
sed "s/%%WINDOWS64_SHA%%/${WINDOWS64_SHA}/" |
sed "s/%%WINDOWS64_SIZE%%/${WINDOWS64_SIZE}/"

# call the tool with something like
# ./linux64/updater -flasher firmwares/NINA/FirmwareUpdater.mkrwifi1010.ino.bin -firmware firmwares/NINA/1.2.1/NINA_W102.bin -port /dev/ttyACM0  -address arduino.cc:443 -restore_binary /tmp/arduino_build_619137/WiFiSSLClient.ino.bin -programmer {runtime.tools.bossac}/bossac
