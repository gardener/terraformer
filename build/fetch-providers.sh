#!/bin/bash -e
#
# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

TF_VERSION="$(cat ./TF_VERSION)"
function version { echo "$@" | awk -F. '{ printf("%d%03d%03d%03d\n", $1,$2,$3,$4); }'; } # https://apple.stackexchange.com/a/123408/11374

terraform-bundle \
  package \
  -os=linux \
  -arch=amd64 \
  <(cat ./terraform-bundle.hcl | sed "s/TF_VERSION/$TF_VERSION/g")

BUNDLE="$(ls -t terraform*.zip | head -1)"
unzip -n $BUNDLE

if [ "$(version "$TF_VERSION")" -lt "$(version "0.13.0")" ]; then
  mkdir "tfproviders"
  find . -name "terraform-provider*" | xargs -I % cp % ./tfproviders/
else
  # Beginning with Terraform 0.13, plugins must be installed in a file system with the following hierarchy:
  # registry.terraform.io/<username>/<plugin-name>/<version>/<arch>
  # Above `terraform-bundle` command already maintains such structure, hence, we keep it instead of copying
  # the `terraform-provider` binaries.
  find . -name "*.md" | xargs -I % rm -f %
  if [ -d ./plugins ]; then
    mv ./plugins ./tfproviders/
  else
    mkdir "tfproviders"
  fi
fi
