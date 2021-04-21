source = ["dist/macos64/FirmwareUploader"]
bundle_id = "cc.arduino.FirmwareUploader"

sign {
  application_identity = "Developer ID Application: ARDUINO SA (7KT7ZWMCJT)"
}

# Ask Gon for zip output to force notarization process to take place.
# The CI will ignore the zip output, using the signed binary only.
zip {
  output_path = "FirmwareUploader.zip"
}