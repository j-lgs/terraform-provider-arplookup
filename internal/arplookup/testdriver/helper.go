package testdriver

import (
	"bytes"
	"fmt"
	"os/exec"
)

// runCmds runs a list of commands in order, returning an error with it's output if execution of any fails.
func runCmds(cmds []*exec.Cmd) error {
	for _, cmd := range cmds {
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running command \"%s\": %w: %s", cmd.String(), err, out.String())
		}
	}

	return nil
}
