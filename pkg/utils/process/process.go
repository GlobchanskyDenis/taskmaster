package process

import (
	// "syscall"
	"os/exec"
	"io"
	// "os"
)

func New(processName, processDir string, args, env []string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	cmd := exec.Command(processName)
	cmd.Dir = processDir
	cmd.Args = append([]string{processName}, args...)
	cmd.Env = env
	/*	Для перехвата стандартных потоков вывода и ошибок  */
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}
	return cmd, stdOut, stdErr, nil
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