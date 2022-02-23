package process

import (
	"testing"
	"os/exec"
	"io"
	"fmt"
	"bufio"
)

func TestStdout(t *testing.T) {
	cmd, stdout, stderr, err := New("../../../sample_simple_bin", "./", nil, nil)
	if err != nil {
		t.Errorf("Error %s", err)
		t.FailNow()
	}
	
	defer close(t, cmd, stdout, stderr)

	fmt.Println("cmd pid ", cmd.Process.Pid)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Printf("==%s==\n", scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("Error %s", err)
	}
}

func close(t *testing.T, cmd *exec.Cmd, stdout, stderr io.ReadCloser) {
	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Error while process kill: %s", err)
	}
	if err := stdout.Close(); err != nil {
		t.Errorf("Error while stdout close: %s", err)
	}
	if err := stderr.Close(); err != nil {
		t.Errorf("Error while stderr close: %s", err)
	}
}