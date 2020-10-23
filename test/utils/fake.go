// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
