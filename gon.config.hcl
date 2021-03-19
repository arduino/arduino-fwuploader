source = ["dist/macos64/FirmwareUpdater"]
bundle_id = "cc.arduino.FirmwareUpdater"

sign {
  application_identity = "Developer ID Application: ARDUINO SA (7KT7ZWMCJT)"
}

# Ask Gon for zip output to force notarization process to take place.
# The CI will ignore the zip output, using the signed binary only.
zip {
  output_path = "FirmwareUpdater.zip"
}