package invoker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CmdInvokerTestSuite is a test suite for CmdInvoker
type CmdInvokerTestSuite struct {
	suite.Suite
}

// TestInvokeSuccess tests the success case of invoking a command
func (suite *CmdInvokerTestSuite) TestInvokeSuccess() {
	ctx := context.Background()
	invoker := NewCmdInvoker()
	exitCode, output, err := invoker.Invoke(ctx, "echo", "hello world")

	// assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, exitCode)
	assert.Equal(suite.T(), "hello world\n", output)
}

// TestInvokeFailure tests the failure case of invoking a command that does not exist
func (suite *CmdInvokerTestSuite) TestInvokeFailure() {
	ctx := context.Background()
	invoker := NewCmdInvoker()
	exitCode, output, err := invoker.Invoke(ctx, "non-existent-command")

	// assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), -1, exitCode)
	assert.Equal(suite.T(), "", output)
}

// TestInvokeFailureTimeout tests the failure case of invoking a command that times out
func (suite *CmdInvokerTestSuite) TestInvokeFailureTimeout() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelFunc()
	invoker := NewCmdInvoker()
	exitCode, output, err := invoker.Invoke(ctx, "bash", "-c", "echo hello world && sleep 1 && echo goodbye world")

	// assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), -1, exitCode)
	assert.Equal(suite.T(), "hello world\n", output)
}

func TestCmdInvokerTestSuite(t *testing.T) {
	suite.Run(t, new(CmdInvokerTestSuite))
}
