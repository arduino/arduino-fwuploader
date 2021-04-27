package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/arduino/FirmwareUploader/modules/nina"
	"github.com/arduino/FirmwareUploader/modules/sara"
	"github.com/arduino/FirmwareUploader/modules/winc"
	"github.com/arduino/FirmwareUploader/utils"
	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/go-paths-helper"
)

var ctx = &context.Context{}

func init() {
	flag.StringVar(&ctx.PortName, "port", "", "serial port to use for flashing")
	flag.StringVar(&ctx.RootCertDir, "certs", "", "root certificate directory")
	flag.Var(&ctx.Addresses, "address", "address (host:port) to fetch and flash root certificate for, multiple values allowed")
	flag.StringVar(&ctx.FirmwareFile, "firmware", "", "firmware file to flash")
	flag.BoolVar(&ctx.ReadAll, "read", false, "read all firmware and output to stdout")
	flag.StringVar(&ctx.FWUploaderBinary, "flasher", "", "firmware upload binary (precompiled for the right target)")
	flag.StringVar(&ctx.BinaryToRestore, "restore_binary", "", "binary to restore after the firmware upload (precompiled for the right target)")
	flag.StringVar(&ctx.ProgrammerPath, "programmer", "", "path of programmer in use (avrdude/bossac)")
	flag.StringVar(&ctx.Model, "model", "", "module model (winc, nina or sara)")
	flag.StringVar(&ctx.Compatible, "get_available_for", "", "Ask for available firmwares matching a given board")
	flag.IntVar(&ctx.Retries, "retries", 9, "Number of retries in case of upload failure")
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

	if ctx.BinaryToRestore != "" {
		// sanity check for BinaryToRestore
		f := paths.New(ctx.BinaryToRestore)
		info, err := f.Stat()
		if err != nil {
			log.Fatalf("Error opening restore_binary: %s", err)
		}
		if info.IsDir() {
			log.Fatalf("Error opening restore_binary: is a directory...")
		}
		if info.Size() == 0 {
			log.Println("WARNING: restore_binary is empty! Will not restore binary after upload.")
			ctx.BinaryToRestore = ""
		}
	}

	retry := 0
	for {
		var ctxCopy context.Context
		ctxCopy = *ctx
		var err error
		if ctx.Model == "nina" || strings.Contains(ctx.FirmwareFile, "NINA") || strings.Contains(ctx.FWUploaderBinary, "NINA") {
			err = nina.Run(&ctxCopy)
		} else if ctx.Model == "winc" || strings.Contains(ctx.FirmwareFile, "WINC") || strings.Contains(ctx.FWUploaderBinary, "WINC") {
			err = winc.Run(&ctxCopy)
		} else {
			err = sara.Run(&ctxCopy)
		}
		if err == nil {
			log.Println("Operation completed: success! :-)")
			break
		}
		log.Println("Error: " + err.Error())

		if retry >= ctx.Retries {
			log.Fatal("Operation failed. :-(")
		}

		retry++
		log.Println("Waiting 1 second before retrying...")
		time.Sleep(time.Second)
		log.Printf("Retrying upload (%d of %d)", retry, ctx.Retries)
	}
}
