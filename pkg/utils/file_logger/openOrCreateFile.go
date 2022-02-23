package file_logger

import (
	"path/filepath"
	"strconv"
	"os"
)

func openOrCreateNewFile(pathToFile, fileName, logFileCreatePem string) (*os.File, error) {
	FilePem, err := strFilePemToUint(logFileCreatePem)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(pathToFile); err != nil {
		err := os.MkdirAll(pathToFile, os.FileMode(FilePem))
		if err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(filepath.Join(pathToFile, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.FileMode(FilePem))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func strFilePemToUint(parseString string) (uint32, error) {
	ui32, err := strconv.ParseUint(parseString, 8, 32)
	if err != nil {
		return 0, err
	}
	return uint32(ui32), nil
}