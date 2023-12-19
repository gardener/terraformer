# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    aws = {
      version = "4.55.0"
    }
    azurerm = {
      version = "3.44.0"
    }
    google = {
      version = "4.53.1"
    }
    google-beta = {
      version = "4.53.1"
    }
    openstack = {
      version = "1.49.0"
      source   = "terraform-provider-openstack/openstack"
    }
    alicloud = {
      version = "1.213.0"
    }
    metal = {
      version = "3.1.0"
      source   = "equinix/metal"
    }
    template = {
      version = "2.1.2"
    }
    null = {
      version = "3.2.1"
    }
  }
}
