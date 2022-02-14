package osProcess

import (
	"syscall"
	"os"
)

func New(processName, processDir string, args, env []string, parentStopSignal syscall.Signal) (*os.Process, error) {
	return os.StartProcess(processName, args, &os.ProcAttr{
		Dir: processDir,
		Env: env,
		Files: nil,
		Sys: &syscall.SysProcAttr{
			Pdeathsig: parentStopSignal,
		},
	})
}

func GetByPid(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}