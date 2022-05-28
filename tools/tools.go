//go:build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "honnef.co/go/tools/cmd/staticcheck"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
