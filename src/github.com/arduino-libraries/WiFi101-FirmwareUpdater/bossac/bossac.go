package bossac

import (
	"bytes"
	"errors"
	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/context"
	serial "go.bug.st/serial.v1"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	//"go.bug.st/serial.v1/enumerator"
	"time"
)

type Bossac struct {
}

func (b *Bossac) Flash(ctx *context.Context, filename string) error {
	log.Println("Flashing " + filename)

	port, err := reset(ctx.PortName, true)
	if err != nil {
		return err
	}
	err = invokeBossac([]string{ctx.ProgrammerPath, "-e", "-R", "-p", port, "-w", filename})

	ports, err := serial.GetPortsList()
	ctx.PortName = waitPort(ports, port)

	return err
}

func (b *Bossac) DumpAndFlash(ctx *context.Context, filename string) (string, error) {
	log.Println("Flashing " + filename)
	dir, err := ioutil.TempDir("", "wifiFlasher_dump")
	port, err := reset(ctx.PortName, true)
	if err != nil {
		return "", err
	}
	err = invokeBossac([]string{ctx.ProgrammerPath, "-u", "-r", "-p", port, filepath.Join(dir, "dump.bin")})
	log.Println("Original sketch saved at " + filepath.Join(dir, "dump.bin"))
	if err != nil {
		return "", err
	}
	err = invokeBossac([]string{ctx.ProgrammerPath, "-e", "-R", "-p", port, "-w", filename})

	ports, err := serial.GetPortsList()
	ctx.PortName = waitPort(ports, port)

	return filepath.Join(dir, "dump.bin"), err
}

func invokeBossac(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	log.Println(out.String())
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

// reset opens the port at 1200bps. It returns the new port name (which could change
// sometimes) and an error (usually because the port listing failed)
func reset(port string, wait bool) (string, error) {
	log.Println("Restarting in bootloader mode")

	// Get port list before reset
	ports, err := serial.GetPortsList()
	log.Println("Get port list before reset")
	if err != nil {
		return "", errors.New("Get port list before reset")
	}

	// Touch port at 1200bps
	err = touchSerialPortAt1200bps(port)
	if err != nil {
		return "", errors.New("1200bps Touch")
	}

	// Wait for port to disappear and reappear
	if wait {
		port = waitReset(ports, port)
	}

	return port, nil
}

// waitReset is meant to be called just after a reset. It watches the ports connected
// to the machine until a port disappears and reappears. The port name could be different
// so it returns the name of the new port.
func waitReset(beforeReset []string, originalPort string) string {
	var port string
	timeout := false

	go func() {
		time.Sleep(10 * time.Second)
		timeout = true
	}()

	for {
		ports, _ := serial.GetPortsList()
		port = differ(ports, beforeReset)

		if port != "" {
			break
		}
		if timeout {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Wait for the port to reappear
	log.Println("Wait for the port to reappear")
	afterReset, _ := serial.GetPortsList()
	for {
		ports, _ := serial.GetPortsList()
		port = differ(ports, afterReset)
		if port != "" {
			time.Sleep(time.Millisecond * 500)
			break
		}
		if timeout {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	// try to upload on the existing port if the touch was ineffective
	if port == "" {
		port = originalPort
	}

	return port
}

func waitPort(beforeReset []string, originalPort string) string {
	var port string
	timeout := false

	go func() {
		time.Sleep(10 * time.Second)
		timeout = true
	}()

	for {
		ports, _ := serial.GetPortsList()
		port = differ(ports, beforeReset)

		if port != "" {
			break
		}
		if timeout {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	// try to upload on the existing port if the touch was ineffective
	if port == "" {
		port = originalPort
	}

	return port
}

// differ returns the first item that differ between the two input slices
func differ(slice1 []string, slice2 []string) string {
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			return mKey
		}
	}

	return ""
}
