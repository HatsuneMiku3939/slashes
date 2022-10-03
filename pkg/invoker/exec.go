package invoker

import (
	"context"
	"os/exec"
)

// CmdInvoker is a Command Invoker implementation.
type CmdInvoker struct {
}

// New returns a new Command Invoker instance.
func NewCmdInvoker() Invoker {
	return &CmdInvoker{}
}

// Invoke invokes the command in a child process and returns the exit code with console outputs.
func (i *CmdInvoker) Invoke(ctx context.Context, command string, args ...string) (int, string, error) {
	// create a command
	cmd := exec.CommandContext(ctx, command, args...)

	// run the command
	out, err := cmd.CombinedOutput()
	if err != nil {
		return cmd.ProcessState.ExitCode(), string(out), err
	}

	// return the exit code and console outputs
	return cmd.ProcessState.ExitCode(), string(out), nil
}
