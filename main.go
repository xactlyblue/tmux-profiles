package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func execTmuxCommand(args ...string) ([]byte, error) {
	// var stdout, stderr bytes.Buffer
	var output bytes.Buffer

	cmd := exec.Command("tmux", args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		if output.Len() > 0 {
			return nil, fmt.Errorf("run command (%v): %v", err, output.String())
		}

		return nil, fmt.Errorf("run command: %v", err)
	}

	return output.Bytes(), nil
}

func main() {
	if _, err := execTmuxCommand("new", "-d", "-s", "testsession"); err != nil {
		fmt.Printf("Failed to run tmux: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Created new TMUX session\n")
	}
}
