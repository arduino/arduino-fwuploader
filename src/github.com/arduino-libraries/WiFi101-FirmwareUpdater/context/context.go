package context

type addressFlags []string

func (af *addressFlags) String() string {
	return ""
}

func (af *addressFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}

type Context struct {
	PortName 			string
	RootCertDir 	string
	Addresses 		addressFlags
	FirmwareFile 	string
	FWUploaderBinary string
	ReadAll 			bool
	BinaryToRestore string
}
