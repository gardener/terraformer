# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  version = "TF_VERSION"
}

providers {
  google = {
    versions = ["4.53.1"]
  }
  google-beta = {
    versions = ["4.53.1"]
  }
  template = {
    versions = ["2.1.2"]
  }
  null = {
    versions = ["2.1.2"]
  }
}
