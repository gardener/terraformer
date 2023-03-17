# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    metal = {
      version = "3.1.0"
      source = "equinix/metal"
    }
  }
}
