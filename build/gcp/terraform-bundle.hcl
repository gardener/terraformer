# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    google = {
      version = "4.53.1"
    }
    google-beta = {
      version = "4.53.1"
    }
    null = {
      version = "3.2.1"
    }
  }
}
