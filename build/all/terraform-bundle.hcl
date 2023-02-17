# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  version = "TF_VERSION"
}

providers {
  aws = {
    versions = ["4.55.0"]
  }
  azurerm = {
    versions = ["3.44.0"]
  }
  google = {
    versions = ["4.53.1"]
  }
  google-beta = {
    versions = ["4.53.1"]
  }
  openstack = {
    versions = ["1.49.0"]
    source = "terraform-provider-openstack/openstack"
  }
  alicloud = {
    versions = ["1.149.0"]
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
