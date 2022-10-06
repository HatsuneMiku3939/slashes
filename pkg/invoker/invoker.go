package invoker

//go:generate mockery --all

import (
	"context"
)

// Invoker provides an interface for invoking a command in child processes.
// It invokes the command in a child process and returns the exit code with console outputs.
type Invoker interface {
	Invoke(ctx context.Context, command string, args ...string) (int, string, error)
}
