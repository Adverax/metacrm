package access

type readerWriter struct {
	GetterSetter
	Reader
	Writer
}

// NewReaderWriter is constructor for build ReaderWriter, based on the GetterSetter
func NewReaderWriter(gs GetterSetter) ReaderWriter {
	return &readerWriter{
		GetterSetter: gs,
		Reader:       &reader{Getter: gs},
		Writer:       &writer{Setter: gs},
	}
}
