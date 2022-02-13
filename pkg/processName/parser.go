package processName

import (
	"path/filepath"
	"errors"
	"strings"
)

func Parse(rawName string) (name string, path string, args []string, err error) {
	cleanedRawName := strings.Trim(rawName, " ")
	if cleanedRawName == "" {
		return "", "", nil, errors.New("Имя процесса не может быть пустым")
	}
	nameParts := strings.Split(strings.Trim(rawName, " "), " ")
	
	/*	Нахожу аргументы  */
	if len(nameParts) > 1 {
		args = nameParts[1:]
	}

	/*	Нахожу путь процесса (он может быть как задан так и не задан) */
	path, name = filepath.Split(nameParts[0])

	/*	Если путь не задан - присваиваю ему значение '/bin/'  */
	if path == "" {
		path = "/bin/"
	}

	return name, path, args, nil
}