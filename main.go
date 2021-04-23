package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/arduino/FirmwareUploader/modules/nina"
	"github.com/arduino/FirmwareUploader/modules/sara"
	"github.com/arduino/FirmwareUploader/modules/winc"
	"github.com/arduino/FirmwareUploader/utils"
	"github.com/arduino/FirmwareUploader/utils/context"
)

var ctx = &context.Context{}

func init() {
	flag.StringVar(&ctx.PortName, "port", "", "serial port to use for flashing")
	flag.StringVar(&ctx.RootCertDir, "certs", "", "root certificate directory")
	flag.Var(&ctx.Addresses, "address", "address (host:port) to fetch and flash root certificate for, multiple values allowed")
	flag.StringVar(&ctx.FirmwareFile, "firmware", "", "firmware file to flash")
	flag.BoolVar(&ctx.ReadAll, "read", false, "read all firmware and output to stdout")
	flag.StringVar(&ctx.FWUploaderBinary, "flasher", "", "firmware upload binary (precompiled for the right target)")
	flag.StringVar(&ctx.BinaryToRestore, "restore_binary", "", "firmware upload binary (precompiled for the right target)")
	flag.StringVar(&ctx.ProgrammerPath, "programmer", "", "path of programmer in use (avrdude/bossac)")
	flag.StringVar(&ctx.Model, "model", "", "module model (winc, nina or sara)")
	flag.StringVar(&ctx.Compatible, "get_available_for", "", "Ask for available firmwares matching a given board")
}

func main() {
	flag.Parse()

	if ctx.Compatible != "" {
		el, _ := json.Marshal(utils.GetCompatibleWith(ctx.Compatible, ""))
		fmt.Println(string(el))
		os.Exit(0)
	}

	if ctx.PortName == "" {
		log.Fatal("Please specify a serial port")
	}

	if ctx.Model == "nina" || strings.Contains(ctx.FirmwareFile, "NINA") || strings.Contains(ctx.FWUploaderBinary, "NINA") {
		nina.Run(ctx)
	} else if ctx.Model == "winc" || strings.Contains(ctx.FirmwareFile, "WINC") || strings.Contains(ctx.FWUploaderBinary, "WINC") {
		winc.Run(ctx)
	} else {
		sara.Run(ctx)
	}
}
