// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package binary_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/gardener/terraformer/pkg/terraformer"
	testutils "github.com/gardener/terraformer/test/utils"
)

var _ = Describe("terraformer", func() {
	var (
		writer                   io.Writer
		pathToTerraformer        string
		terraformerBuildArgsHash string
	)

	buildTerraformer := func(buildArgs ...string) {
		By("building terraformer")
		if pathToTerraformer != "" && testutils.CanReuseLastBuild("terraformer", &terraformerBuildArgsHash, buildArgs) {
			return
		}

		var err error
		pathToTerraformer, err = gexec.Build(
			testutils.TerraformerPackage+"/cmd/terraformer",
			buildArgs...,
		)
		Expect(err).ShouldNot(HaveOccurred())
	}

	BeforeEach(func() {
		writer = gexec.NewPrefixedWriter("terraformer: ", GinkgoWriter)
	})

	Context("terraform not installed", func() {
		BeforeEach(func() {
			buildTerraformer(
				testutils.LDFlags(testutils.OverwriteTerraformBinary("non-existing"))...,
			)
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
			baseDir            string
			paths              *terraformer.PathSet
			fakeTerraform      testutils.FakeTerraform
			overwriteExitCodes testutils.Overwrite

			testObjs *testutils.TestObjects
			args     []string
		)

		BeforeEach(func() {
			var err error
			baseDir, err = ioutil.TempDir("", "tf-test-*")
			Expect(err).NotTo(HaveOccurred())
			paths = terraformer.DefaultPaths().WithBaseDir(baseDir)

			var handle testutils.CleanupActionHandle
			handle = testutils.AddCleanupAction(func() {
				defer testutils.RemoveCleanupAction(handle)
				Expect(os.RemoveAll(baseDir)).To(Succeed())
			})
			overwriteExitCodes = testutils.OverwriteExitCode("0")

			buildTerraformer(
				testutils.LDFlags(testutils.OverwriteTerraformBinary(testutils.FakeTerraformBinary))...,
			)

			testObjs = testutils.PrepareTestObjects(ctx, testClient)
			args = []string{"--zap-devel=true", "--base-dir=" + baseDir}
		})

		JustBeforeEach(func() {
			fakeTerraform = testutils.NewFakeTerraform(
				testutils.OverwriteSleepDuration("50ms"),
				overwriteExitCodes,
			)
		})

		It("should fail, if no command is given", func() {
			cmd := exec.Command(pathToTerraformer, args...)
			cmd.Env = fakeTerraform.TerraformerEnv()

			session, err := gexec.Start(cmd, writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("no command was specified"))
		})

		It("should fail, if --configuration-configmap-name is not set", func() {
			args = append(args,
				"apply",
			)
			cmd := exec.Command(pathToTerraformer, args...)
			cmd.Env = fakeTerraform.TerraformerEnv()

			session, err := gexec.Start(cmd, writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("flag --configuration-configmap-name was not set"))
		})

		It("should fail, if --state-configmap-name is not set", func() {
			args = append(args,
				"apply",
				"--configuration-configmap-name="+testObjs.ConfigurationConfigMap.Name,
			)
			cmd := exec.Command(pathToTerraformer, args...)
			cmd.Env = fakeTerraform.TerraformerEnv()

			session, err := gexec.Start(cmd, writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("flag --state-configmap-name was not set"))
		})

		It("should fail, if --variables-secret-name is not set", func() {
			args = append(args,
				"apply",
				"--configuration-configmap-name="+testObjs.ConfigurationConfigMap.Name,
				"--state-configmap-name="+testObjs.StateConfigMap.Name,
			)
			cmd := exec.Command(pathToTerraformer, args...)
			cmd.Env = fakeTerraform.TerraformerEnv()

			session, err := gexec.Start(cmd, writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("flag --variables-secret-name was not set"))
		})

		It("should fail, if no kubeconfig has been provided", func() {
			args = append(args,
				"apply",
				"--configuration-configmap-name="+testObjs.ConfigurationConfigMap.Name,
				"--state-configmap-name="+testObjs.StateConfigMap.Name,
				"--variables-secret-name="+testObjs.VariablesSecret.Name,
			)
			cmd := exec.Command(pathToTerraformer, args...)
			cmd.Env = fakeTerraform.TerraformerEnv()

			session, err := gexec.Start(cmd, writer, writer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Eventually(session.Err).Should(gbytes.Say("no configuration has been provided"))
		})

		Context("successful terraform execution", func() {
			It("should correctly fetch config and state from API server", func() {
				args = append(args,
					"apply",
					"--namespace="+testObjs.Namespace,
					"--configuration-configmap-name="+testObjs.ConfigurationConfigMap.Name,
					"--state-configmap-name="+testObjs.StateConfigMap.Name,
					"--variables-secret-name="+testObjs.VariablesSecret.Name,
					"--kubeconfig="+kubeconfigFile,
				)
				cmd := exec.Command(pathToTerraformer, args...)
				cmd.Env = fakeTerraform.TerraformerEnv()

				session, err := gexec.Start(cmd, writer, writer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 3).Should(gexec.Exit(0))

				contents, err := ioutil.ReadFile(filepath.Join(paths.ConfigDir, testutils.ConfigMainKey))
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(BeEquivalentTo(testObjs.ConfigurationConfigMap.Data[testutils.ConfigMainKey]))
				contents, err = ioutil.ReadFile(filepath.Join(paths.ConfigDir, testutils.ConfigVarsKey))
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(BeEquivalentTo(testObjs.ConfigurationConfigMap.Data[testutils.ConfigVarsKey]))

				contents, err = ioutil.ReadFile(paths.StatePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(BeEquivalentTo(testObjs.StateConfigMap.Data[testutils.StateKey]))

				contents, err = ioutil.ReadFile(paths.VarsPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(BeEquivalentTo(testObjs.VariablesSecret.Data[testutils.VarsKey]))
			})
		})
	})
})
