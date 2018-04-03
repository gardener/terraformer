#!/bin/bash -eu
#
# Copyright 2017 The Gardener Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIR_STATE_CONFIG_MAP="/tf-state-in"
DIR_STATE_IN="/tf-state-copy"
DIR_STATE_OUT="/tf-state-out"
DIR_VARIABLES="/tfvars"
DIR_PROVIDERS="/terraform-providers"

PATH_STATE_CONFIG_MAP="$PATH_STATE_CONFIG_MAP"
PATH_STATE_IN="$DIR_STATE_IN/terraform.tfstate"
PATH_STATE_OUT="$DIR_STATE_OUT/terraform.tfstate"
PATH_VARIABLES="$DIR_VARIABLES/terraform.tfvars"

PATH_CACERT="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
NAMESPACE"$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
TOKEN="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
BASE_URL="https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT_HTTPS"

# determine command
command="${1:-apply}"
exitcode=1

mkdir -p "$DIR_STATE_IN"
mkdir -p "$DIR_STATE_OUT"

# required to initialize the AWS provider plugin (since v0.10.0, see https://www.terraform.io/upgrade-guides/0-10.html)
terraform init -plugin-dir="$DIR_PROVIDERS" /tf
# workaround for `terraform init`; required to make `terraform validate` work (plugin_path file ignored?)
cp -r "$DIR_PROVIDERS"/* .terraform/plugins/linux_amd64/.
# copy input terraform state into another directory because it's mounted as configmap (read-only filesystem) and terraform will try to write onto that fs
cp "$DIR_STATE_CONFIG_MAP"/* "$DIR_STATE_IN"

function end_execution() {
  # Delete trap handler to avoid recursion
  trap - HUP QUIT PIPE INT TERM EXIT

  # check whether the terraform state has changed
  if [[ ! -f "$PATH_STATE_OUT" ]] || diff "$PATH_STATE_IN" "$PATH_STATE_OUT" 1> /dev/null; then # passes (returns 0) if there is no diff
    # indicate success (exit code gets lost as the surrounding command pipes this script through 'tee')
    echo -e "\nNothing to do."
    if [[ $exitcode -eq 0 ]]; then
      touch /success
    fi
    exit 0
  else
    # update config map with the new terraform state (see https://stackoverflow.com/questions/30690186/how-do-i-access-the-kubernetes-api-from-within-a-pod-container)
    echo -e "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: $TF_STATE_CONFIG_MAP_NAME\ndata:\n  terraform.tfstate: |" > "$PATH_STATE_CONFIG_MAP"
    cat "$PATH_STATE_OUT" | sed -n 's/^/    /gp' >> "$PATH_STATE_CONFIG_MAP"
    curl \
      --silent \
      --cacert "$PATH_CACERT" \
      --header "Authorization: Bearer $TOKEN" \
      --header "Content-Type: application/yaml" \
      --request PUT \
      --data-binary @"$PATH_STATE_CONFIG_MAP" \
      "$BASE_URL/api/v1/namespaces/$NAMESPACE/configmaps/$TF_STATE_CONFIG_MAP_NAME" > /dev/null

    # validate that the current terraform state is properly reflected in the config map
    curl \
      --silent \
      --cacert "$PATH_CACERT" \
      --header "Authorization: Bearer $TOKEN" \
      --header "Accept: application/yaml" \
      --request GET \
      "$BASE_URL/api/v1/namespaces/$NAMESPACE/configmaps/$TF_STATE_CONFIG_MAP_NAME" > "$PATH_STATE_CONFIG_MAP.put"

    sed -i -n 's/^    //gp' "$DIR_STATE_OUT/$TF_STATE_CONFIG_MAP_NAME.put"
    if diff "$PATH_STATE_OUT" "$DIR_STATE_OUT/$TF_STATE_CONFIG_MAP_NAME.put" 1> /dev/null; then # passes (returns 0) if there is no diff
      # indicate success (exit code gets lost as the surrounding command pipes this script through 'tee')
      echo -e "\nConfigMap successfully updated with terraform state."
      if [[ $exitcode -eq 0 ]]; then
        touch /success
      fi
      exit 0
    else
      # dump terraform state so that we can find it at least in the logs
      echo -e "\nConfigMap could not be updated with terraform state! Dumping state file now to have it in the logs:"
      cat "$PATH_STATE_OUT"
      exit 1
    fi
  fi
}

if [[ "$command" == "validate" ]]; then
  terraform \
    validate \
    -check-variables=true \
    -var-file="$PATH_VARIABLES" \
    /tf
  if [[ "$?" == "0" ]]; then
    terraform \
      plan \
      -parallelism=4 \
      -detailed-exitcode \
      -state="$PATH_STATE_IN" \
      -var-file="$PATH_VARIABLES" \
      /tf
  else
    exit $?
  fi
else
  # Install trap handler for proper cleanup, summary and result code
  trap "exit 100" HUP QUIT PIPE
  trap end_execution INT TERM EXIT
  if [[ "$command" == "apply" ]]; then
    terraform \
      apply \
      -auto-approve \
      -parallelism=4 \
      -state="$PATH_STATE_IN" \
      -state-out="$PATH_STATE_OUT" \
      -var-file="$PATH_VARIABLES" \
      /tf
    exitcode=$?
  elif [[ "$command" == "destroy" ]]; then
    terraform \
      destroy \
      -force \
      -parallelism=4 \
      -state="$PATH_STATE_IN" \
      -state-out="$PATH_STATE_OUT" \
      -var-file="$PATH_VARIABLES" \
      /tf
    exitcode=$?
  else
    # Delete trap handler - nothing to do
    trap - HUP INT QUIT PIPE TERM EXIT
    echo -e "\nUnknown command: $command"
    exit 1
  fi
fi
