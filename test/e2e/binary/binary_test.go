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

package binary_test

import (
	"io"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	testutils "github.com/gardener/terraformer/test/utils"
)

var _ = Describe("terraformer", func() {
	var (
		writer            io.Writer
		pathToTerraformer string
	)

	BeforeSuite(func() {
		writer = gexec.NewPrefixedWriter("terraformer: ", GinkgoWriter)
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("terraform not installed", func() {
		BeforeEach(func() {
			var err error

			By("building terraformer")
			pathToTerraformer, err = gexec.Build(
				testutils.TerraformerPackage+"/cmd/terraformer",
				testutils.LDFlags(testutils.OverwriteTerraformBinary("non-existing"))...,
			)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should fail, if terraform is not installed", func() {
			session, err := gexec.Start(exec.Command(pathToTerraformer), writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("terraform is not installed"))
		})
	})

	Context("fake terraformer installed", func() {
		var (
			fakeTerraform testutils.FakeTerraform
		)

		BeforeEach(func() {
			var err error

			By("building fake terraform")
			fakeTerraform = testutils.NewFakeTerraform(
				testutils.OverwriteExitCode("1"),
				testutils.OverwriteSleepDuration("1s"),
			)

			By("building terraformer")
			pathToTerraformer, err = gexec.Build(
				testutils.TerraformerPackage+"/cmd/terraformer",
				testutils.LDFlags(testutils.OverwriteTerraformBinary(testutils.FakeTerraformBinary))...,
			)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should fail, if no command is given", func() {
			cmd := exec.Command(pathToTerraformer, "--zap-devel=true")
			cmd.Env = fakeTerraform.TerraformerEnv()

			session, err := gexec.Start(cmd, writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("no command was specified"))
		})
	})
})
