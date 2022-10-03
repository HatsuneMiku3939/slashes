//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/spf13/cobra-cli"
	_ "github.com/vektra/mockery/v2"
	_ "golang.org/x/tools/cmd/godoc"
)
