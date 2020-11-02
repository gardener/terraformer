// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package terraformer_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/terraformer/pkg/terraformer"
	testutils "github.com/gardener/terraformer/test/utils"
)

var _ = Describe("Terraformer Config", func() {
	var (
		tf       *terraformer.Terraformer
		paths    *terraformer.PathSet
		testObjs *testutils.TestObjects
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

		tf, err = terraformer.NewTerraformer(
			&terraformer.Config{
				Namespace:                  testObjs.Namespace,
				ConfigurationConfigMapName: testObjs.ConfigurationConfigMap.Name,
				StateConfigMapName:         testObjs.StateConfigMap.Name,
				VariablesSecretName:        testObjs.VariablesSecret.Name,
				RESTConfig:                 restConfig,
			},
			runtimelog.Log,
			paths,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		testutils.RunCleanupActions()
	})

	Describe("#FetchConfigAndState", func() {
		BeforeEach(func() {
			Expect(tf.EnsureTFDirs()).To(Succeed())
		})

		It("should fetch all objects if present", func() {
			Expect(tf.FetchConfigAndState(ctx)).To(Succeed())

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

		Context("config fetching", func() {
			It("should fail if config ConfigMap is not present", func() {
				Expect(testClient.Delete(ctx, testObjs.ConfigurationConfigMap)).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(MatchError(ContainSubstring("not found")))
			})
			It("should fail if main config key is not present", func() {
				delete(testObjs.ConfigurationConfigMap.Data, testutils.ConfigMainKey)
				Expect(testClient.Update(ctx, testObjs.ConfigurationConfigMap)).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(MatchError(ContainSubstring("not found")))
			})
			It("should fail if vars config key is not present", func() {
				delete(testObjs.ConfigurationConfigMap.Data, testutils.ConfigVarsKey)
				Expect(testClient.Update(ctx, testObjs.ConfigurationConfigMap)).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(MatchError(ContainSubstring("not found")))
			})
		})

		Context("variables fetching", func() {
			It("should fail if variables secret is not present", func() {
				Expect(testClient.Delete(ctx, testObjs.VariablesSecret)).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(MatchError(ContainSubstring("not found")))
			})
			It("should fail if vars key is not present", func() {
				delete(testObjs.VariablesSecret.Data, testutils.VarsKey)
				Expect(testClient.Update(ctx, testObjs.VariablesSecret)).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(MatchError(ContainSubstring("not found")))
			})
		})

		Context("state fetching", func() {
			It("should create empty state file if state ConfigMap is not present", func() {
				Expect(client.IgnoreNotFound(testClient.Delete(ctx, testObjs.StateConfigMap))).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(Succeed())

				contents, err := ioutil.ReadFile(paths.StatePath)
				Expect(err).NotTo(HaveOccurred(), "state file should be present")
				Expect(contents).To(BeEmpty(), "state file should be empty")
			})
			It("should truncate existing state file if state ConfigMap is not present", func() {
				Expect(ioutil.WriteFile(paths.StatePath, []byte("state from last run"), 0644)).To(Succeed())

				Expect(client.IgnoreNotFound(testClient.Delete(ctx, testObjs.StateConfigMap))).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(Succeed())

				contents, err := ioutil.ReadFile(paths.StatePath)
				Expect(err).NotTo(HaveOccurred(), "state file should be present")
				Expect(contents).To(BeEmpty(), "state file should be empty")
			})
			It("should create empty state file if state key is not present", func() {
				delete(testObjs.StateConfigMap.Data, testutils.StateKey)
				Expect(testClient.Update(ctx, testObjs.StateConfigMap)).To(Succeed())

				Expect(tf.FetchConfigAndState(ctx)).To(Succeed())

				contents, err := ioutil.ReadFile(paths.StatePath)
				Expect(err).NotTo(HaveOccurred(), "state file should be present")
				Expect(contents).To(BeEmpty(), "state file should be empty")
			})
		})
	})

	Describe("#StoreState", func() {
		var (
			storeCtx    context.Context
			storeCancel context.CancelFunc
		)

		BeforeEach(func() {
			storeCtx, storeCancel = context.WithTimeout(ctx, 1*time.Minute)
			Expect(tf.EnsureTFDirs()).To(Succeed())
		})

		AfterEach(func() {
			storeCancel()
		})

		It("should store contents of state file in state ConfigMap", func() {
			stateContents := "state from last run"
			Expect(ioutil.WriteFile(paths.StatePath, []byte(stateContents), 0644)).To(Succeed())

			Expect(tf.StoreState()).To(Succeed())

			testObjs.Refresh()
			Expect(testObjs.StateConfigMap.Data).To(HaveKeyWithValue(testutils.StateKey, stateContents))
		})
		It("should store contents of state file in new state ConfigMap", func() {
			Expect(testClient.Delete(ctx, testObjs.StateConfigMap)).To(Succeed())

			stateContents := "state from last run"
			Expect(ioutil.WriteFile(paths.StatePath, []byte(stateContents), 0644)).To(Succeed())

			Expect(tf.StoreState()).To(Succeed())

			testObjs.Refresh()
			Expect(testObjs.StateConfigMap.Data).To(HaveKeyWithValue(testutils.StateKey, stateContents))
		})
	})

	Describe("#StartFileWatcher", func() {
		var (
			fileWatcherCtx    context.Context
			fileWatcherCancel context.CancelFunc
			wg                sync.WaitGroup
		)

		BeforeEach(func() {
			fileWatcherCtx, fileWatcherCancel = context.WithTimeout(ctx, 1*time.Minute)
			Expect(tf.EnsureTFDirs()).To(Succeed())
			Expect(tf.FetchConfigAndState(ctx)).To(Succeed())
		})

		AfterEach(func() {
			fileWatcherCancel()
			wg.Wait()
		})

		It("should not update state ConfigMap if file is not changed", func() {
			Expect(tf.StartFileWatcher(fileWatcherCtx, &wg)).To(Succeed())

			stateBefore := testObjs.StateConfigMap.DeepCopy()

			Consistently(func() runtime.Object {
				testObjs.Refresh()
				return testObjs.StateConfigMap
			}, 1, 0.1).Should(DeepEqual(stateBefore))
		})

		It("should not update state ConfigMap if file is deleted", func() {
			Expect(tf.StartFileWatcher(fileWatcherCtx, &wg)).To(Succeed())

			stateBefore := testObjs.StateConfigMap.DeepCopy()

			Expect(os.Remove(paths.StatePath)).To(Succeed())

			Consistently(func() runtime.Object {
				testObjs.Refresh()
				return testObjs.StateConfigMap
			}, 1, 0.1).Should(DeepEqual(stateBefore))
		})

		It("should update state ConfigMap if file is changed", func() {
			Expect(tf.StartFileWatcher(fileWatcherCtx, &wg)).To(Succeed())

			By("update state file the first time")
			stateContents := "state, generation 1"
			Expect(ioutil.WriteFile(paths.StatePath, []byte(stateContents), 0644)).To(Succeed())

			Eventually(func() map[string]string {
				testObjs.Refresh()
				return testObjs.StateConfigMap.Data
			}, 1, 0.1).Should(HaveKeyWithValue(testutils.StateKey, stateContents))

			By("update state file a second time")
			stateContents = "state, generation 2"
			Expect(ioutil.WriteFile(paths.StatePath, []byte(stateContents), 0644)).To(Succeed())

			Eventually(func() map[string]string {
				testObjs.Refresh()
				return testObjs.StateConfigMap.Data
			}, 1, 0.1).Should(HaveKeyWithValue(testutils.StateKey, stateContents))
		})
	})
})
