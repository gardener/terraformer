// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package binary_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// This test suite runs e2e tests against the terraformer binary. It builds the binary using gexec.Build and optionally
// builds a terraformer fake binary to mock terraform executions in terraformer.
// It starts a fake control plane (via envtest) and doesn't require an existing cluster by default, so it can be run in
// normal CI pipelines.
// Nevertheless, an existing cluster can still be used like in the unit tests by setting the KUBECONFIG env var and
// setting USE_EXISTING_CLUSTER=true.
func TestTerraformer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terraformer Binary E2E Suite")
}
