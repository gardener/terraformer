// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package terraformer_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/gardener/gardener/pkg/utils/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gardener/terraformer/pkg/terraformer"
	"github.com/gardener/terraformer/pkg/utils"
	testutils "github.com/gardener/terraformer/test/utils"
)

var _ = Describe("Terraformer", func() {
	Describe("#NewDefaultTerraformer", func() {
		It("should fail, if it can't create a client", func() {
			_, err := terraformer.NewDefaultTerraformer(&terraformer.Config{})
			Expect(err).To(MatchError(ContainSubstring("kubernetes client")))
		})
	})

	Describe("#Run", func() {
		var (
			tf *terraformer.Terraformer
		)

		BeforeEach(func() {
			var err error
			tf, err = terraformer.NewDefaultTerraformer(&terraformer.Config{RESTConfig: restConfig})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail, if command is not supported", func() {
			Expect(tf.Run("non-existing")).To(MatchError(ContainSubstring("not supported")))
		})
		It("should not allow to run Init directly", func() {
			Expect(tf.Run(terraformer.Init)).To(MatchError(ContainSubstring("not supported")))
		})
		It("should not allow to run Plan directly", func() {
			Expect(tf.Run(terraformer.Plan)).To(MatchError(ContainSubstring("not supported")))
		})
	})

	Describe("#Run", func() {
		var (
			fakeTerraform testutils.FakeTerraform
			tf            *terraformer.Terraformer
			paths         *terraformer.PathSet
			testObjs      *testutils.TestObjects

			resetVars              func()
			testStdout, testStderr *gbytes.Buffer
		)

		BeforeEach(func() {
			baseDir, err := ioutil.TempDir("", "tf-test-*")

			var handle testutils.CleanupActionHandle
			handle = testutils.AddCleanupAction(func() {
				defer testutils.RemoveCleanupAction(handle)
				Expect(os.RemoveAll(baseDir)).To(Succeed())
			})
			Expect(err).NotTo(HaveOccurred())

			paths = terraformer.DefaultPaths().WithBaseDir(baseDir)

			testObjs = testutils.PrepareTestObjects(ctx, testClient)

			testStdout = gbytes.NewBuffer()
			testStderr = gbytes.NewBuffer()

			resetVars = test.WithVars(
				&terraformer.Stdout, io.MultiWriter(GinkgoWriter, testStdout),
				&terraformer.Stderr, io.MultiWriter(GinkgoWriter, testStderr),
			)

			tf, err = terraformer.NewTerraformer(
				&terraformer.Config{
					Namespace:                  testObjs.Namespace,
					ConfigurationConfigMapName: testObjs.ConfigurationConfigMap.Name,
					StateConfigMapName:         testObjs.StateConfigMap.Name,
					VariablesSecretName:        testObjs.VariablesSecret.Name,
					RESTConfig:                 restConfig,
				},
				zap.New(zap.UseDevMode(true), zap.WriteTo(io.MultiWriter(GinkgoWriter, testStdout))),
				paths,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			resetVars()
		})

		Context("successful terraform execution", func() {
			var (
				resetBinary func()
			)

			BeforeEach(func() {
				By("building fake terraform")
				fakeTerraform = testutils.NewFakeTerraform(
					testutils.OverwriteExitCode("0"),
					testutils.OverwriteSleepDuration("50ms"),
				)

				resetBinary = test.WithVars(
					&terraformer.TerraformBinary, fakeTerraform.Path,
				)
			})

			AfterEach(func() {
				resetBinary()
			})

			It("should run Apply successfully", func() {
				Expect(tf.Run(terraformer.Apply)).To(Succeed())
				Eventually(testStdout).Should(gbytes.Say("some terraform output"))
				Eventually(testStdout).Should(gbytes.Say("terraform process finished successfully"))
			})
			It("should run Destroy successfully", func() {
				Expect(tf.Run(terraformer.Destroy)).To(Succeed())
				Eventually(testStdout).Should(gbytes.Say("some terraform output"))
				Eventually(testStdout).Should(gbytes.Say("terraform process finished successfully"))
			})
			It("should run Validate successfully", func() {
				Expect(tf.Run(terraformer.Validate)).To(Succeed())
				Eventually(testStdout).Should(gbytes.Say("some terraform output"))
				Eventually(testStdout).Should(gbytes.Say("terraform process finished successfully"))
			})
		})

		Context("failed terraform execution", func() {
			var (
				resetBinary func()
			)

			BeforeEach(func() {
				By("building fake terraform")
				fakeTerraform = testutils.NewFakeTerraform(
					testutils.OverwriteExitCode("42"),
					testutils.OverwriteSleepDuration("50ms"),
				)

				resetBinary = test.WithVars(
					&terraformer.TerraformBinary, fakeTerraform.Path,
				)
			})

			AfterEach(func() {
				resetBinary()
			})

			It("should return exit code from terraform", func() {
				err := tf.Run(terraformer.Apply)
				Expect(err).To(MatchError(ContainSubstring("terraform command failed")))

				var withExitCode utils.WithExitCode
				Expect(errors.As(err, &withExitCode)).To(BeTrue())
				Expect(withExitCode.ExitCode()).To(Equal(42))

				Eventually(testStderr).Should(gbytes.Say("some terraform error"))
				Eventually(testStdout).Should(gbytes.Say("terraform process finished with error"))
				Eventually(testStdout).Should(gbytes.Say("storing state before exiting"))
				Eventually(testStdout).Should(gbytes.Say("successfully stored terraform state"))
			})
		})
	})
})
