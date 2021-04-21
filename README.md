# Firmware/Certificates updater for WINC and NINA Wifi module

Use this tool to update the firmware and/or add SSL certificates for any WINC, NINA or SARA module.

## Install

You can download the Firmware/Certificates updater here:

https://github.com/arduino/FirmwareUploader/releases/latest

## Usage

Extract the zip file and run (for example, NINA -> WiFi1010)

```
./FirmwareUploader -flasher firmwares/NINA/FirmwareUpdater.mkrwifi1010.ino.bin -firmware firmwares/NINA/1.2.1/NINA_W102.bin -port /dev/ttyACM0 -address arduino.cc:443 -restore_binary /tmp/arduino_build_619137/WiFiSSLClient.ino.bin -programmer {runtime.tools.bossac}/bossac
```

To flash a MKR1000:

```
./FirmwareUploader -flasher firmwares/WINC1500/FirmwareUpdater.mkr1000.ino.bin -firmware firmwares/WINC1500/19.5.4/m2m_aio_3a0.bin -port /dev/ttyACM0 -address arduino.cc:443 -restore_binary /tmp/arduino_build_619137/WiFiSSLClient.ino.bin -programmer {runtime.tools.bossac}/bossac
```

To update a MKRNB1500:

```
./FirmwareUploader -flasher firmwares/SARA/SerialSARAPassthrough.ino.bin -firmware firmwares/SARA/5.6A2.00-to-5.6A2.01.pkg -port /dev/ttyACM0 -restore_binary firmwares/SARA/SerialSARAPassthrough.ino.bin -programmer {runtime.tools.bossac}/bossac
```

### Command line options

The full list of command line options can be obtained with the `-h` option: `./FirmwareUploader -h`

```
Usage of ./FirmwareUploader:
  -address value
      address (host:port) to fetch and flash root certificate for, multiple values allowed
  -certs string
      root certificate directory
  -firmware string
      firmware file to flash
  -flasher string
      firmware upload binary (precompiled for the right target)
  -get_available_for string
      Ask for available firmwares matching a given board
  -model string
      module model (winc, nina or sara)
  -port string
      serial port to use for flashing
  -programmer string
      path of programmer in use (avrdude/bossac)
  -read
      read all firmware and output to stdout
  -restore_binary string
      firmware upload binary (precompiled for the right target)
```

## How to build the tools from source file

To build we use [task](https://taskfile.dev/) for simplicity. From the sources root directory run:

```
task dist:<OS>_<ARCH>
```

Where <OS> could be one of: `macOS`,`Windows`,`Linux`. And <ARCH>: `32bit`, `64bit`, `ARM` or `ARM64`

This will create the `FirmwareUploader` executable.

## Security

If you think you found a vulnerability or other security-related bug in this project, please read our [security
policy][security-policy] and report the bug to our Security Team 🛡️ Thank you!

e-mail contact: security@arduino.cc

## License

Copyright (c) 2015-2021 Arduino LLC. All right reserved.

This library is free software; you can redistribute it and/or modify it under the terms of the GNU Lesser General Public
License as published by the Free Software Foundation; either version 2.1 of the License, or (at your option) any later
version.

This library is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
details.

You should have received a copy of the GNU Lesser General Public License along with this library; if not, write to the
Free Software Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA

[security-policy]: https://github.com/arduino/FirmwareUploader/security/policy
