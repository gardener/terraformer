// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import "fmt"

const (
	// TerraformerPackage is the terraformer package path
	TerraformerPackage = "github.com/gardener/terraformer"
)

// Overwrite is a variable overwrite which can be passed to -ldflags
type Overwrite struct {
	VarPath, Value string
}

func (o Overwrite) String() string {
	return fmt.Sprintf("-X %s=%s", o.VarPath, o.Value)
}

// LDFlags returns ldflags to pass to `go build` or `gexec.Build` in tests
func LDFlags(overwrites ...Overwrite) []string {
	if len(overwrites) == 0 {
		return []string{}
	}

	var ldflags string
	for _, o := range overwrites {
		ldflags += " " + o.String()
	}
	return []string{"-ldflags", ldflags}
}

// OverwriteTerraformBinary returns an overwrite for the terraform binary name
func OverwriteTerraformBinary(path string) Overwrite {
	return Overwrite{
		VarPath: TerraformerPackage + "/pkg/terraformer.TerraformBinary",
		Value:   path,
	}
}

// OverwriteExitCode returns an overwrite for the exit code var
func OverwriteExitCode(code string) Overwrite {
	return Overwrite{
		VarPath: "main.exitCode",
		Value:   code,
	}
}

// OverwriteSleepDuration returns an overwrite for the exit code var
func OverwriteSleepDuration(code string) Overwrite {
	return Overwrite{
		VarPath: "main.sleepDuration",
		Value:   code,
	}
}
