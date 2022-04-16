package file_writer

import (
	"os"
	"io"
)

func New(pathfile string) io.WriteCloser {
	file, err := os.OpenFile(pathfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err == nil && file != nil {
		return file
	}

	return nil
}
