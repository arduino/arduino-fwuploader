package context

type Context struct {
	PortName         string
	RootCertDir      string
	Addresses        []string
	FirmwareFile     string
	FWUploaderBinary string
	ReadAll          bool
	BinaryToRestore  string
	ProgrammerPath   string
	Model            string
	BoardName        string
	Retries          int
}
