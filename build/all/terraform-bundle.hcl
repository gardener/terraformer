# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  version = "TF_VERSION"
}

providers {
  aws         = ["3.32.0"]
  azurerm     = ["2.36.0"]
  google      = ["3.62.0"]
  google-beta = ["3.62.0"]
  openstack   = ["1.37.0"]
  alicloud    = ["1.124.2"]
  packet      = ["2.3.0"]
  template    = ["2.1.2"]
  null        = ["2.1.2"]
}
