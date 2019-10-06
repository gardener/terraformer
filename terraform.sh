#!/bin/bash -eu
#
# Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.
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

PATH_STATE_CONFIG_MAP="$DIR_STATE_OUT/$TF_STATE_CONFIG_MAP_NAME"
PATH_STATE_IN="$DIR_STATE_IN/terraform.tfstate"
PATH_STATE_OUT="$DIR_STATE_OUT/terraform.tfstate"
PATH_VARIABLES="$DIR_VARIABLES/terraform.tfvars"

NAMESPACE="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"

# exponential backoff functionality
: ${MAX_TIME_SEC:=1800}
: ${MAX_BACKOFF_SEC:=30}

function backoff() {
  t=1
  start="$(date +%s)"

  while ! "${@}"; do
    now="$(date +%s)"
    if [[ $((start + MAX_TIME_SEC)) -le $now ]]; then
      echo "$(date) Timeout of ${MAX_TIME_SEC}s reached. Aborting"
      return 1
    fi

    echo "$(date) Repeating command in $t sec..."
    sleep $t

    t=$((t * 2))
    if [[ "$t" -ge "${MAX_BACKOFF_SEC}" ]]; then
      t=${MAX_BACKOFF_SEC}
    fi
  done

  echo "$(date) Command executed successful."
}

# determine command
command="${1:-apply}"
exitcode=1

mkdir -p "$DIR_STATE_IN"
mkdir -p "$DIR_STATE_OUT"

# required to initialize the provider plugins
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
    # update config map with the new terraform state
    echo -e "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: $TF_STATE_CONFIG_MAP_NAME\ndata:\n  terraform.tfstate: |" > "$PATH_STATE_CONFIG_MAP"
    cat "$PATH_STATE_OUT" | sed -n 's/^/    /gp' >> "$PATH_STATE_CONFIG_MAP"

    function update_state_configmap() {
      kubectl \
        replace \
        -f "$PATH_STATE_CONFIG_MAP" \
        > /dev/null

      # validate that the current terraform state is properly reflected in the config map
      kubectl \
        --namespace="$NAMESPACE" \
        get configmap "$TF_STATE_CONFIG_MAP_NAME" \
        --output="jsonpath={.data['terraform\.tfstate']}" \
        > "$PATH_STATE_CONFIG_MAP.put"

      if diff "$PATH_STATE_OUT" "$PATH_STATE_CONFIG_MAP.put" 1> /dev/null; then # passes (returns 0) if there is no diff
        # indicate success (exit code gets lost as the surrounding command pipes this script through 'tee')
        echo -e "\nConfigMap successfully updated with terraform state."
        if [[ $exitcode -eq 0 ]]; then
          touch /success
        fi
        return 0
      else
        return 1
      fi
    }

    if ! backoff update_state_configmap; then
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
      -auto-approve \
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
