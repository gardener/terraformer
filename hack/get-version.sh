#!/bin/bash -e
#
# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

VERSION="$(cat "$(dirname $0)/../VERSION")"

if [[ "$VERSION" = *-dev ]] ; then
  VERSION="$VERSION-$(git rev-parse HEAD)"
fi

# .dockerignore ignores all files unrelevant for build (e.g. docs) to only copy relevant source files to the build
# container. Hence, git will always detect a dirty work tree when building in a container (many deleted files).
# This command filters out all deleted files that are ignored by .dockerignore to only detect changes to relevant files
# as a dirty work tree.
# Additionally, it filters out changes to the `VERSION` file, as this is currently the only way to inject the
# version-to-build in our pipelines (see https://github.com/gardener/cc-utils/issues/431).
TREE_STATE="$([ -z "$(git status --porcelain 2>/dev/null | grep -vf <(git ls-files --deleted --ignored --exclude-from=.dockerignore) -e 'VERSION')" ] && echo clean || echo dirty)"

if [ "$TREE_STATE" = dirty ] ; then
  VERSION="$VERSION-dirty"
fi

echo "$VERSION"
