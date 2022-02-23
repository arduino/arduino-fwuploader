## Usage

### Firmware Flashing

First [install] the Arduino Firmware Uploader. Extract the zip file and, for example, to update the NINA module present
on the Arduino MKR WiFi 1010, run:

```
./arduino-fwuploader firmware flash -b arduino:samd:mkrwifi1010 -a /dev/ttyACM0
```

You just have to specify the FQBN (`-b` or `--fqbn`) of the board and the serial port (`-a` or `--address`). The
firmware uploader will take care of fetching everything is required to perform the update process. If no module and
version are specified **the latest version of the firmware** will be used.

If you want to flash a specific version of a firmware you can use the `-m` or `--module` flag

For example to flash the WINC module present on the MKR 1000 with version 19.6.1 of the firmware you can run something
like:

```
./arduino-fwuploader firmware flash -b arduino:samd:mkr1000 -a /dev/ttyACM0 -m WINC1500@19.6.1
```

Or if you want upload a local firmware you can use the `-i` or `--input-file` flag followed by the path of that
firmware.

```
./arduino-fwuploader firmware flash -i custom_fw.bin -b arduino:samd:mkr1000 -a /dev/ttyACM0
```

There is a retry mechanism because the flashing process uses serial communication, which sometimes can be a bit
unreliable. The retry flag is set by default to 9 retries, but it's possible to overwrite it for whatever reason. For
example to update a Nano RP2040 Connect with the retry set to 2 you can use:

```
./arduino-fwuploader firmware flash --fqbn arduino:mbed_nano:nanorp2040connect -a /dev/ttyACM0 --retries 2
```

It's possible to list the available firmwares for every board/module with:

```
./arduino-fwuploader firmware list
```

but you can also filter the results by specifying the `-b` or `--fqbn` flag

The tool offers the ability to print output in JSON, with the `--format json` flag

### Get Version

You can also obtain the version of the firmware the board is currently running with:

```
./arduino-fwuploader firmware get-version -b arduino:samd:mkrwifi1010 -a /dev/ttyACM0
```

The `get-version` subcommand flashes a special sketch in order to be able to read that information using the serial
connection:

```
...

Firmware version installed: 1.4.8
```

You can also use the `--format json` flag to parse the output with more ease.

### Certificates

The tool offers also the ability to flash SSL certificates to a module:

```
./arduino-fwuploader certificates flash -b arduino:samd:nano_33_iot" -a COM10 -u arduino.cc:443 -u google.com:443
```

or you can specify a path to a file in `.der` format with `-f` instead of the URL of the certificate

Due to a limitation on the handling of certs in the NINA modules, when `certificates flash` command is launched, all the
previous certificates are going to be erased. To overcome this it's required to upload them all together:

```
./arduino-fwuploader certificates flash --url arduino.cc:443,google.com:443 --file ~/Downloads/my_self_signed_.der -b arduino:mbed_nano:nanorp2040connect -a /dev/ttyACM0
```

### Command line options

The full list of command line options can be obtained with the `-h` option: `./arduino-fwuploader -h`

For further information you can use the [command reference]

[install]: installation.md
[command reference]: commands/arduino-fwuploader.md
