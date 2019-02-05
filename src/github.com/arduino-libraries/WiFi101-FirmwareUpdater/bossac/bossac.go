package bossac

import (
	"bytes"
  "errors"
  "path/filepath"
  "log"
  "os/exec"
	"io/ioutil"
	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/context"
  serial "go.bug.st/serial.v1"
	//"go.bug.st/serial.v1/enumerator"
  "time"
  "os"
)

func Restore(ctx context.Context, filename string) error {
  err := touchSerialPortAt1200bps(ctx.PortName)
  if err != nil {
    return err
  }
  err = invokeBossac([]string{"-e", "-w" , filename})
  os.RemoveAll(filepath.Dir(filename))
  return err
}

func DumpAndFlash(ctx context.Context) (string, error) {
  dir, err := ioutil.TempDir("wifiFlasher", "dump")
  err = touchSerialPortAt1200bps(ctx.PortName)
  if err != nil {
    return "", err
  }
  err = invokeBossac([]string{"-r", filepath.Join(dir, "dump.bin")})
  if err != nil {
    return "", err
  }
  err = invokeBossac([]string{"-e", "-w" , ctx.FlasherBinary})
  return filepath.Join(dir, "dump.bin"), err
}

func invokeBossac(args []string) error {
  cmd := exec.Command("bossac/bossac", args...)
  var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return err
}

func touchSerialPortAt1200bps(port string) error {
	log.Println("Touching port " + port + " at 1200bps")

	// Open port
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return errors.New("Open port " + port)
	}
	defer p.Close()

	// Set DTR
	err = p.SetDTR(false)
	log.Println("Set DTR off")
	if err != nil {
		return errors.New("Can't set DTR")
	}

	// Wait a bit to allow restart of the board
	time.Sleep(200 * time.Millisecond)

	return nil
}
