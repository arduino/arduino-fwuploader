# Firmware/Certificates updater for WINC and NINA Wifi module

Use this tool to update the firmware and/or add SSL certificates for any WINC, NINA or SARA module.

## Install

You can download the Firmware/Certificates updater here:

https://github.com/arduino/FirmwareUploader/releases/latest

## Usage

### Firmware Flashing

Extract the zip file and, to update a mkr 1010, run:

```
./arduino-fwuploader firmware flash -b arduino:samd:mkrwifi1010 -a /dev/ttyACM0
```

You just have to specify the fqbn (`-b` or `--fqbn`) of the board and the serial port (`-a` or `--address`) The firmware
uploader will take care of fetching everything is required to perform the update process. If no module and version are
specified **the latest version of the firmware** will be used.

If you want to flash a specific version of a firmware you can use the `-m` or `--module` flag

For example to flash a MKR1000 with 19.6.1 version of the firmware you can run something like:

```
./arduino-fwuploader firmware flash -b arduino:samd:mkr1000 -a /dev/ttyACM0 -m WINC1500@19.6.1
```

There is also a retry mechanism bundled in the tool because the flashing process sometimes can be a bit unreliable. For
example to update a Nano RP2040 Connect with the retry set to 2 you can use:

```
./arduino-fwuploader firmware flash --fqbn arduino:mbed_nano:nanorp2040connect -a /dev/ttyACM0 --retries 2
```

It's possible also to list the available firmwares for every board/module with:

```
./arduino-fwuploader firmware list
```

but you can also filter the results by specifying the `-b` or `--fqbn` flag

The tool offers the ability to print output in json, with the `--format json`

### Certificates

The tool offers also the ability to flash SSL certificates to a board:

```
/arduino-fwuploader flash -b arduino:samd:nano_33_iot" -a COM10 -u arduino.cc:443 -u google.cc:443
```

or you can specify a path to a file with `-f`

### Command line options

The full list of command line options can be obtained with the `-h` option: `./arduino-fwuploader -h`

```
Arduino Firmware Uploader (arduino-fwuploader).

Usage:
  arduino-fwuploader [command]

Examples:
  ./arduino-fwuploader <command> [flags...]

Available Commands:
  certificates Commands to operate on certificates.
  firmware     Commands to operate on firmwares.
  help         Help about any command
  version      Shows version number of arduino-fwuploader.

Flags:
      --format string       The output format, can be {text|json}. (default "text")
  -h, --help                help for arduino-fwuploader
      --log-file string     Path to the file where logs will be written
      --log-format string   The output format for the logs, can be {text|json}.
      --log-level string    Messages with this level and above will be logged. Valid levels are: trace, debug, info, warn, error, fatal, panic (default "info")
  -v, --verbose             Print the logs on the standard output.

Use "arduino-fwuploader [command] --help" for more information about a command.
```

## How to build the tools from source file

To build we use [task](https://taskfile.dev/) for simplicity. From the sources root directory run:

```
task dist:<OS>_<ARCH>
```

Where <OS> could be one of: `macOS`,`Windows`,`Linux`. And <ARCH>: `32bit`, `64bit`, `ARM` or `ARM64`

This will create the `arduino-fwuploader` executable.

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
