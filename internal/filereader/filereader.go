package filereader

import (
	log "github.com/sirupsen/logrus"
	"os"
)

type File interface {
	Read(p []byte) (n int, err error)
	Stat() (os.FileInfo, error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}

type Reader interface {
	Open(name string) (File, error)
}

type fileReaderImpl struct {
	logger *log.Entry
}

func New(logger *log.Entry) Reader {
	return &fileReaderImpl{
		logger: logger,
	}
}

// Open opens a file for reading.
func (d *fileReaderImpl) Open(name string) (File, error) {
	return os.Open(name)
}
