package main

// import (
// 	"bytes"
// 	"fmt"
// 	"os/exec"
// )

// func main() {
//   cmd := exec.Command("vim")
//   var buf bytes.Buffer
//   cmd.Stdout = &buf
//   err := cmd.Start()
//   if err != nil {
//     fmt.Printf("error: %v\n", err)
//   }
//   err = cmd.Wait()

//   fmt.Printf("Command finished with error: %v\n", err)
//   fmt.Printf("Command finished with output: %v\n", buf.String())
// }

import (
	"syscall"
	"fmt"
	"os"
)

func main() {
  process, err := os.StartProcess("vim", []string{}, &os.ProcAttr{
	  Dir: "/bin/",
	  Env: []string{},
	  Files: []*os.File{},
	  Sys: &syscall.SysProcAttr{},
  })
//   var buf bytes.Buffer
//   cmd.Stdout = &buf
//   err := cmd.Start()
//   if err != nil {
//     fmt.Printf("error: %v\n", err)
//   }
//   err = cmd.Wait()

	if err != nil {
		fmt.Printf("Command finished with error: %s\n", err)
	} else {
		fmt.Printf("Command finished with output: %#v\n", process)
	}
}
