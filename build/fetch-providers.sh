#!/bin/bash -e
#
# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

mkdir "tfproviders"
terraform providers mirror tfproviders

BUNDLE="$(ls -t tfproviders/terraform*.zip | head -1)"
unzip -n $BUNDLE

find . -name "terraform-provider*" | xargs -I % cp % ./tfproviders/

