package osProcess

import "os"

func New(processName, processDir string, args, env []string) (*os.Process, error) {
	return os.StartProcess(processName, args, &os.ProcAttr{
		Dir: processDir,
		Env: env,
		Files: nil,
		Sys: nil,
	})
}

func GetByPid(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}