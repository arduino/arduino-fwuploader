## Install

You can download the Arduino Firmware Uploader directly from the GitHub release page:

https://github.com/arduino/arduino-fwuploader/releases/latest

### How to build the tools from source file

To build we use [task](https://taskfile.dev/) for simplicity. From the sources root directory run:

```
task dist:<OS>_<ARCH>
```

Where <OS> could be one of: `macOS`,`Windows`,`Linux`. And <ARCH>: `32bit`, `64bit`, `ARMv6`, `ARMv7` or `ARM64`

This will create the `arduino-fwuploader` executable.
