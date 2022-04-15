package file_writer

import (
	"os"
	"io"
)

func New(pathfile string) io.WriteCloser {
	// file, err := os.Open(pathfile)
	// if err == nil && file != nil {
	// 	return file
	// }

	_ = os.Remove(pathfile)

	file, err := os.Create(pathfile)
	if err == nil && file != nil {
		return file
	}

	return nil
}
