package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/process"
	"fmt"
)

func main() {
	proc, err := process.GetByPid(31149)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		if err := proc.Kill(); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}

	proc, err = process.GetByPid(30932)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		if err := proc.Kill(); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
}
