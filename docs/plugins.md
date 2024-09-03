# Plugins

A new plugin-system has been implemented in the Arduino Firmware Uploader. A plugin is basically another
executable/tool, specifically tailored for a board/family of boards, but with a well-defined user interface. Inside the
plugin, there is all the business logic to support a board/family of boards. Every plugin implements the
[interface](https://pkg.go.dev/github.com/arduino/fwuploader-plugin-helper#Plugin) contained in the
[fwuploader-plugin-helper](https://github.com/arduino/fwuploader-plugin-helper) repository.

The fwuploader is still responsible for downloading the files required for the plugin to work:

- plugins/tools
- certificates
- firmware

The new information is hosted in the
[plugin_firmware_index.json](https://downloads.arduino.cc/arduino-fwuploader/boards/plugin_firmware_index.json), which
is generated by `generator.py`.

## Portenta C33

The source code can be found [here](https://github.com/arduino/portenta-c33-fwuploader-plugin).

## Nina family

The source code can be found [here](https://github.com/arduino/nina-fwuploader-plugin). Supported boards:

- [MKR WiFi 1010](https://docs.arduino.cc/hardware/mkr-wifi-1010)
- [Nano 33 IoT](https://docs.arduino.cc/hardware/nano-33-iot)
- [UNO WiFi Rev2](https://docs.arduino.cc/hardware/uno-wifi-rev2)
- [Nano RP2040 Connect](https://docs.arduino.cc/hardware/nano-rp2040-connect)

## UNO R4 WiFi

The source code can be found [here](https://github.com/arduino/uno-r4-wifi-fwuploader-plugin).

### Known issues

#### Espflash panic `UnknownModel`

On some arm64 Linux distros, version 2.0.0 of [espflash](https://github.com/esp-rs/espflash/) might panic with the
following error:

```
Error:   × Main thread panicked.
  ├─▶ at espflash/src/interface.rs:70:33
  ╰─▶ called `Result::unwrap()` on an `Err` value: UnknownModel
  help: set the `RUST_BACKTRACE=1` environment variable to display a
        backtrace.
```

#### The ESP32 module does not go into download mode

On Linux, the UNO R4 WiFi must be plugged into a **USB hub** to make the flash process work. Otherwise, it won’t be able
to reboot in download mode.

```bash
$ arduino-fwuploader firmware flash -b arduino:renesas_uno:unor4wifi -a /dev/ttyACM0 -v --log-level debug

Done in 0.001 seconds
Write 46588 bytes to flash (12 pages)
[==============================] 100% (12/12 pages)
Done in 3.106 seconds

Waiting to flash the binary...
time=2023-07-18T14:50:10.492+02:00 level=INFO msg="getting firmware version"
time=2023-07-18T14:50:10.509+02:00 level=INFO msg="firmware version is > 0.1.0 using sketch"
time=2023-07-18T14:50:10.511+02:00 level=INFO msg="check if serial port has changed"
[2023-07-18T12:50:20Z INFO ] 🚀 A new version of espflash is available: v2.0.1
[2023-07-18T12:50:20Z INFO ] Serial port: '/dev/ttyACM0'
[2023-07-18T12:50:20Z INFO ] Connecting...
[2023-07-18T12:50:20Z INFO ] Unable to connect, retrying with extra delay...
[2023-07-18T12:50:21Z INFO ] Unable to connect, retrying with default delay...
[2023-07-18T12:50:21Z INFO ] Unable to connect, retrying with extra delay...
[2023-07-18T12:50:21Z INFO ] Unable to connect, retrying with default delay...
[2023-07-18T12:50:21Z INFO ] Unable to connect, retrying with extra delay...
[2023-07-18T12:50:21Z INFO ] Unable to connect, retrying with default delay...
[2023-07-18T12:50:21Z INFO ] Unable to connect, retrying with extra delay...
Error: espflash::connection_failed

  × Error while connecting to device
  ╰─▶ Failed to connect to the device
  help: Ensure that the device is connected and the reset and boot pins are
        not being held down

Error: exit status 1
ERRO[0021] couldn't update firmware: exit status 3
INFO[0021] Waiting 1 second before retrying...
INFO[0022] Uploading firmware (try 2 of 9)
time=2023-07-18T14:50:22.229+02:00 level=INFO msg=upload_command_sketch
time=2023-07-18T14:50:22.230+02:00 level=INFO msg="sending serial reset"
Error: reboot mode: upload commands sketch: setting DTR to OFF
...
```

#### I flashed the certificates, but I am unable to reach the host

There was a bug in the arduino-fwuploader prior `2.4.1` which didn't pick the actual root certificate. Upgrading to the
latest version solves the problem.

#### My antivirus says that `espflash` is a threat

The binary is not signed [#348](https://github.com/esp-rs/espflash/issues/348), and some antiviruses might complain. If
still doubtful, https://github.com/esp-rs/espflash is open source, and it's possible to double-check the md5 hashes of
the binary and the source code. For more information, you can follow
[this](https://forum.arduino.cc/t/radio-module-firmware-version-0-2-0-is-now-available/1147361/11) forum thread.

#### Not running on armv7

At the moment, we are always downloading the armv6 binaries. Since they are dynamically linked, most likely they are not
going to run on armv7. More infos on the topic:
[here](https://developer.arm.com/documentation/ddi0419/c/Appendices/ARMv7-M-Differences/ARMv6-M-and-ARMv7-M-compatibility).