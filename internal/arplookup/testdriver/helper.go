package testdriver

import (
	"bytes"
	"fmt"
	"os/exec"
)

// runCmds runs a list of commands in order, returning an error with it's output if execution of any fails.
func runCmds(cmds []*exec.Cmd) error {
	for _, cmd := range cmds {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running command \"%s\": %w: \n%s\n===\n%s", cmd.String(), err, stdout.String(), stderr.String())
		}
	}

	return nil
}
