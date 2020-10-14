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
