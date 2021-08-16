// Source: https://github.com/arduino/tooling-project-assets/blob/main/workflow-templates/assets/cobra/docsgen/go.mod
module github.com/arduino/arduino-fwuploader/docsgen

go 1.16

replace github.com/arduino/arduino-fwuploader => ../

require (
	github.com/arduino/arduino-fwuploader v0.0.0
	github.com/spf13/cobra v1.1.3
)
