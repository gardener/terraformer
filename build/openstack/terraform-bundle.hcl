# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    openstack = {
      version = "1.49.0"
      source   = "terraform-provider-openstack/openstack"
    }
    null = {
      version = "3.2.1"
    }
  }
}

