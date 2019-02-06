package main

import (
	"flag"
	"log"
	"strings"
	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/nina"
	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/winc"
	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/context"
)

var ctx context.Context

func init() {
	flag.StringVar(&ctx.PortName, "port", "", "serial port to use for flashing")
	flag.StringVar(&ctx.RootCertDir, "certs", "", "root certificate directory")
	flag.Var(&ctx.Addresses, "address", "address (host:port) to fetch and flash root certificate for, multiple values allowed")
	flag.StringVar(&ctx.FirmwareFile, "firmware", "", "firmware file to flash")
	flag.BoolVar(&ctx.ReadAll, "read", false, "read all firmware and output to stdout")
	flag.StringVar(&ctx.FWUploaderBinary, "flasher", "", "firmware upload binary (precompiled for the right target)")
	flag.StringVar(&ctx.BinaryToRestore, "restore_binary", "", "firmware upload binary (precompiled for the right target)")
	flag.StringVar(&ctx.ProgrammerPath, "programmer", "", "path of programmer in use (avrdude/bossac)")
	flag.StringVar(&ctx.Model, "model", "", "module model (winc or nina)")
}

func main() {
	flag.Parse()

	if ctx.PortName == "" {
		log.Fatal("Please specify a serial port")
	}

	if ctx.Model == "nina" || strings.Contains(ctx.FirmwareFile, "NINA") || strings.Contains(ctx.FWUploaderBinary, "NINA") {
		nina.Run(ctx)
	} else {
		winc.Run(ctx)
	}
}
