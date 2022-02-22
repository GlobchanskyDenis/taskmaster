package dto

import (
	"time"
)

type CommandResult struct {
	Pid        int
	Name       string
	StatusCode uint
	Status     string
	Error      error
	ExitCode   int
	ChangeTime time.Time
}
