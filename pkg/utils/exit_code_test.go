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

package utils_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gardener/terraformer/pkg/utils"
)

var _ = Describe("WithExitCode", func() {
	var (
		exitCode        int
		err, underlying error
	)

	BeforeEach(func() {
		exitCode = 1
		underlying = errors.New("foo")
		err = utils.WithExitCode{
			Code:       exitCode,
			Underlying: underlying,
		}
	})

	It("should return specified exit code and unwrap underlying", func() {
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("with exit code %d", exitCode))))
		Expect(err).To(MatchError(ContainSubstring(underlying.Error())))

		var withExitCode utils.WithExitCode
		Expect(errors.As(err, &withExitCode)).To(BeTrue())
		Expect(withExitCode.ExitCode()).To(Equal(exitCode))
		Expect(withExitCode.Unwrap()).To(Equal(underlying))
	})
})
