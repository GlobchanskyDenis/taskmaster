package dto

import (
	"syscall"
)

type ProcessMeta struct {
	Name          string
	BinPath       string
	Args          []string
	Env           []string
	ProcessPath   *string
	Autostart     bool
	Autorestart   bool
	RestartTimes  *uint
	StopSignal    syscall.Signal
	Exitcodes     []int
	Starttime     uint
	Stoptime      uint
	Umask         int
	OutputDiscard bool
}

func (pm ProcessMeta) IsRestartByRestartTimes() bool {
	if pm.Autorestart == true && pm.RestartTimes != nil {
		return true
	}
	return false
}

func (pm ProcessMeta) IsRestartByExitCodeCheck() bool {
	if pm.Autorestart == true && len(pm.Exitcodes) > 0 {
		return true
	}
	return false
}
