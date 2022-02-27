package main

import (
	// "syscall"
	"os"
)

func createTestFile(name, body string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write([]byte(body)); err != nil {
		return err
	}
	return nil
}

// func setUmask(mask int) {
// 	syscall.Umask(mask)
// }