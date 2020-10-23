// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
