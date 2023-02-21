# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    alicloud = {
      version = "1.149.0"
    }
    template = {
      version = "2.1.2"
    }
    null = {
      version = "3.2.1"
    }
  }
}
