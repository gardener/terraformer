// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
