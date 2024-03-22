#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

ENVTEST_K8S_VERSION=${ENVTEST_K8S_VERSION:-"1.25"}

echo "> Installing envtest tools@${ENVTEST_K8S_VERSION} with setup-envtest if necessary"
if ! command -v setup-envtest &> /dev/null ; then
  >&2 echo "setup-envtest not available"
  exit 1
fi

export KUBEBUILDER_ASSETS="$(setup-envtest use -p path ${ENVTEST_K8S_VERSION})"
echo "using envtest tools installed at '$KUBEBUILDER_ASSETS'"
source "$(dirname "$0")/../vendor/github.com/gardener/gardener/hack/test-integration.env"

"$(dirname "$0")"/../vendor/github.com/gardener/gardener/hack/test-cover.sh "$@"
