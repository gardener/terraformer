# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  version = "TF_VERSION"
}

providers {
  aws = {
    versions = ["3.63.0"]
  }
  azurerm = {
    versions = ["2.68.0"]
  }
  google = {
    versions = ["3.62.0"]
  }
  google-beta = {
    versions = ["3.62.0"]
  }
  openstack = {
    versions = ["1.37.0"]
    source = "terraform-provider-openstack/openstack"
  }
  alicloud = {
    versions = ["1.124.2"]
  }
  metal = {
    versions = ["3.1.0"]
    source = "equinix/metal"
  }
  template = {
    versions = ["2.1.2"]
  }
  null = {
    versions = ["2.1.2"]
  }
}
