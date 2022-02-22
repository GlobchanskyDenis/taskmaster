package process

import (
	// "syscall"
	"os/exec"
	// "os"
)

func New(processName, processDir string, args, env []string) (*exec.Cmd, error) {
	cmd := exec.Command(processName, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// func newDeprecated(processName, processDir string, args, env []string, parentStopSignal syscall.Signal) (*os.Process, error) {
// 	return os.StartProcess(processName, args, &os.ProcAttr{
// 		Dir: processDir,
// 		Env: env,
// 		Files: nil,
// 		Sys: &syscall.SysProcAttr{
// 			Pdeathsig: parentStopSignal,
// 		},
// 	})
// }

// func GetByPid(pid int) (*os.Process, error) {
// 	return os.FindProcess(pid)
// }