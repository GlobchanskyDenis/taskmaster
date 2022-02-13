package dto

import (

)

type Command struct {
	Type uint
}

type CommandResult struct {
	Pid        int
	Name       string
	StatusCode uint
	Status     string
	Error      error
}