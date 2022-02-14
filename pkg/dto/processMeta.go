package dto

import (
	"syscall"
)

type ProcessMeta struct {
	Name         string
	BinPath      string
	Args         []string
	Env          []string
	ProcessPath  *string
	Autostart    bool
	Autorestart  bool
	RestartTimes *uint
	StopSignal   syscall.Signal
	Exitcodes    []int
	Starttime    uint
	Stoptime     uint
}
