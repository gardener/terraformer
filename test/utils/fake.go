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

import (
	"os"
	"path"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	// FakeTerraformBinary is the name of the fake terraform binary
	FakeTerraformBinary = "fake-terraform"
)

// FakeTerraform models the terraform fake binary
type FakeTerraform struct {
	// Path is the path to the fake terraform binary
	Path string
}

// NewFakeTerraform builds a new terraform fake binary, which can be used to mock terraform executions in tests
func NewFakeTerraform(overwrites ...Overwrite) FakeTerraform {
	binPath, err := gexec.Build(
		TerraformerPackage+"/test/utils/fake-terraform",
		LDFlags(overwrites...)...,
	)
	Expect(err).NotTo(HaveOccurred())

	return FakeTerraform{binPath}
}

// TerraformerEnv returns the environment variables to use for executing terraformer, so that it will be able to find
// the fake terraform binary.
func (f FakeTerraform) TerraformerEnv() []string {
	env := os.Environ()

	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + path.Dir(f.Path) + ":" + strings.TrimPrefix(e, "PATH=")
		}
	}

	return env
}
