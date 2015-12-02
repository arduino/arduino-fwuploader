/*
  FirmwareUploader.go - A firmware uploader for the WiFi101 module.
  Copyright (c) 2015 Arduino LLC.  All right reserved.

  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.

  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
*/

package main

import (
	"bytes"
	"certificates"
	"errors"
	"flasher"
	_ "fmt"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	_ "github.com/google/gxui/gxfont"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"
	"go.bug.st/serial"
	_ "io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

var payloadSize uint16
var uploading bool

type Cert struct {
	Label string
	Data  certificates.CertEntry
}

func (c *Cert) String() string {
	return c.Label
}

var downloadedCerts []*Cert

func appMain(driver gxui.Driver) {
	theme := flags.CreateTheme(driver)

	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)

	addLabel := func(text string) {
		label := theme.CreateLabel()
		label.SetText(text)
		layout.AddChild(label)
	}

	addLabel("1. Fetch certificates from websites")
	addLabel("Insert IP or domain name here and press 'Fetch' button")

	urlTextBox := theme.CreateTextBox()
	urlTextBox.SetDesiredWidth(300)

	fetchButton := theme.CreateButton()
	fetchButton.SetText("Fetch")
	fetchButton.SetVisible(false)

	{
		lineLayout := theme.CreateLinearLayout()
		lineLayout.SetDirection(gxui.LeftToRight)
		lineLayout.AddChild(urlTextBox)
		lineLayout.AddChild(fetchButton)
		layout.AddChild(lineLayout)
	}

	statusLabel := theme.CreateLabel()
	statusLabel.SetMultiline(true)
	statusLabel.SetText("")
	layout.AddChild(statusLabel)

	certListAdapter := gxui.CreateDefaultAdapter()
	certList := theme.CreateList()
	certList.SetAdapter(certListAdapter)

	removeButton := theme.CreateButton()
	removeButton.SetText("Remove")

	{
		lineLayout := theme.CreateLinearLayout()
		lineLayout.SetDirection(gxui.LeftToRight)
		lineLayout.AddChild(certList)
		lineLayout.AddChild(removeButton)
		layout.AddChild(lineLayout)
	}

	addLabel("2. Select programmer serial port")

	portListAdapter := gxui.CreateDefaultAdapter()
	portList := theme.CreateList()
	portList.SetAdapter(portListAdapter)

	refreshButton := theme.CreateButton()
	refreshButton.SetText("Refresh List")

	{
		lineLayout := theme.CreateLinearLayout()
		lineLayout.SetDirection(gxui.LeftToRight)
		lineLayout.AddChild(portList)
		lineLayout.AddChild(refreshButton)
		layout.AddChild(lineLayout)
	}

	addLabel("3. Upload certificate to WiFi module")

	uploadButton := theme.CreateButton()
	uploadButton.SetText("Upload certificates")
	layout.AddChild(uploadButton)

	progressStatus := theme.CreateLabel()
	layout.AddChild(progressStatus)

	progressBar := theme.CreateProgressBar()
	size := math.MaxSize
	size.H = 20
	progressBar.SetDesiredSize(size)
	layout.AddChild(progressBar)

	// Business logic

	portSelected := false
	updateUploadButton := func() {
		visible := portSelected && certListAdapter.Count() > 0
		uploadButton.SetVisible(visible)
	}
	updateUploadButton()

	updateDownloadedCerts := func() {
		certListAdapter.SetItems(downloadedCerts)
		updateUploadButton()
	}

	downloadCert := func() {
		if uploading {
			return
		}
		url := urlTextBox.Text()
		if strings.Index(url, ":") == -1 {
			url += ":443"
		}
		data, err := certificates.EntryForAddress(url)
		if err != nil {
			log.Println("Error downloading certificate. " + err.Error())
			statusLabel.SetText("Error downloading certificate. " + err.Error())
			return
		} else {
			statusLabel.SetText("Download successful")
		}
		cert := &Cert{
			Label: url,
			Data:  data,
		}
		downloadedCerts = append(downloadedCerts, cert)
		urlTextBox.SetText("")
		updateDownloadedCerts()
	}

	urlTextBox.OnTextChanged(func([]gxui.TextBoxEdit) {
		isEmpty := (urlTextBox.Text() == "")
		fetchButton.SetVisible(!isEmpty)
	})
	urlTextBox.OnKeyPress(func(event gxui.KeyboardEvent) {
		char := event.Key
		if char == gxui.KeyEnter || char == gxui.KeyKpEnter {
			isEmpty := (urlTextBox.Text() == "")
			if !isEmpty {
				downloadCert()
			}
		}
	})

	fetchButton.OnClick(func(gxui.MouseEvent) {
		downloadCert()
	})

	removeButton.OnClick(func(gxui.MouseEvent) {
		if uploading {
			return
		}
		selected := certList.Selected()
		i := certList.Adapter().ItemIndex(selected)
		downloadedCerts = append(downloadedCerts[:i], downloadedCerts[i+1:]...)
		updateDownloadedCerts()
		removeButton.SetVisible(false)
	})
	certList.OnSelectionChanged(func(gxui.AdapterItem) {
		removeButton.SetVisible(true)
	})
	removeButton.SetVisible(false)

	refreshPortList := func() {
		if uploading {
			return
		}
		if list, err := serial.GetPortsList(); err != nil {
			log.Println("Error fetching serial ports" + err.Error())
		} else {
			portListAdapter.SetItems(list)
		}
	}
	refreshPortList()

	refreshButton.OnClick(func(gxui.MouseEvent) {
		refreshPortList()
		portSelected = false
		updateUploadButton()
	})
	portList.OnSelectionChanged(func(gxui.AdapterItem) {
		portSelected = true
		updateUploadButton()
	})

	updateProgress := func(msg string, percent int) {
		time.Sleep(time.Second)
		driver.CallSync(func() {
			if percent == -1 {
				progressStatus.SetColor(gxui.Red)
				progressBar.SetVisible(false)
			} else if percent == 100 {
				progressStatus.SetColor(gxui.Green)
				progressBar.SetVisible(false)
			} else {
				progressStatus.SetColor(gxui.White)
				progressBar.SetProgress(percent)
				progressBar.SetVisible(true)
			}
			progressStatus.SetText(msg)
		})
	}
	progressBar.SetVisible(false)

	uploadButton.OnClick(func(gxui.MouseEvent) {
		if uploading {
			return
		}
		port := portList.Selected().(string)
		uploading = true
		go uploadCertificates(port, driver, updateProgress)
		log.Println(port)
	})

	updateDownloadedCerts()

	window := theme.CreateWindow(800, 600, "Linear layout")
	window.SetTitle("WINC1500 SSL Certificate updater")
	window.SetScale(flags.DefaultScaleFactor)
	window.AddChild(layout)
	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	window.Relayout()
}

func uploadCertificates(portName string, driver gxui.Driver, updateProgress func(string, int)) {
	defer func() { uploading = false }()

	updateProgress("Connecting to programmer on "+portName, 10)
	programmer, err := flasher.Open(portName)
	if err != nil {
		updateProgress(err.Error(), -1)
		return
	}
	defer programmer.Close()

	updateProgress("Synchronizing with programmer", 20)
	if err := programmer.Hello(); err != nil {
		updateProgress(err.Error(), -1)
		return
	}

	updateProgress("Reading programmer capabilities", 30)
	payloadSize, err = programmer.GetMaximumPayloadSize()
	if err != nil {
		updateProgress(err.Error(), -1)
		return
	}
	if payloadSize < 1024 {
		updateProgress("Programmer reports "+strconv.Itoa(int(payloadSize))+" as maximum payload size (1024 is needed)", -1)
		return
	}

	updateProgress("Converting certificates", 40)
	entries := []certificates.CertEntry{}
	for _, downloadedCert := range downloadedCerts {
		entries = append(entries, downloadedCert.Data)
	}
	convertedCers := certificates.ConvertCertEntries(entries)

	updateProgress("Uploading certificates...", 50)
	CertificatesOffset := 0x4000
	err = flashChunk(programmer, CertificatesOffset, convertedCers)
	if err != nil {
		updateProgress(err.Error(), -1)
		return
	}

	// For debugging puporses
	/*
		if err := ioutil.WriteFile("cert_output.bin", convertedCers, 0644); err != nil {
			updateProgress(err.Error(), -1)
			return
		}
	*/

	updateProgress("Upload completed!", 100)
}

func flashChunk(programmer *flasher.Flasher, offset int, buffer []byte) error {
	chunkSize := int(payloadSize)
	bufferLength := len(buffer)

	if err := programmer.Erase(uint32(offset), uint32(bufferLength)); err != nil {
		return err
	}

	for i := 0; i < bufferLength; i += chunkSize {
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := programmer.Write(uint32(offset+i), buffer[start:end]); err != nil {
			return err
		}
	}

	var flashData []byte
	for i := 0; i < bufferLength; i += chunkSize {
		readLength := chunkSize
		if (i + chunkSize) > bufferLength {
			readLength = bufferLength % chunkSize
		}

		data, err := programmer.Read(uint32(offset+i), uint32(readLength))
		if err != nil {
			return err
		}

		flashData = append(flashData, data...)
	}

	if !bytes.Equal(buffer, flashData) {
		return errors.New("Flash data does not match written!")
	}

	return nil
}

func main() {
	gl.StartDriver(appMain)
}
