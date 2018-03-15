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

# determine command
command="${1:-apply}"
exitcode=1

mkdir /tf-state-out

# required to initialize the AWS provider plugin (since v0.10.0, see https://www.terraform.io/upgrade-guides/0-10.html)
terraform init -plugin-dir=/terraform-providers /tf
# workaround for `terraform init`; required to make `terraform validate` work (plugin_path file ignored?)
cp -r /terraform-providers/* .terraform/plugins/linux_amd64/.
# copy input terraform state into another directory because it's mounted as configmap (read-only filesystem) and terraform will try to write onto that fs
mkdir -p /tf-state-tmp
cp /tf-state-in/* /tf-state-tmp

function end_execution() {
  # Delete trap handler to avoid recursion
  trap - HUP QUIT PIPE INT TERM EXIT

  # check whether the terraform state has changed
  if diff /tf-state-tmp/terraform.tfstate /tf-state-out/terraform.tfstate 1> /dev/null; then # passes (returns 0) if there is no diff
    # indicate success (exit code gets lost as the surrounding command pipes this script through 'tee')
    echo -e "\nNothing to do."
    if [[ $exitcode -eq 0 ]]; then
      touch /success
    fi
    exit 0
  else
    # update config map with the new terraform state (see https://stackoverflow.com/questions/30690186/how-do-i-access-the-kubernetes-api-from-within-a-pod-container)
    echo -e "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: $TF_STATE_CONFIG_MAP_NAME\ndata:\n  terraform.tfstate: |" > /tf-state-out/$TF_STATE_CONFIG_MAP_NAME
    cat /tf-state-out/terraform.tfstate | sed -n 's/^/    /gp' >> /tf-state-out/$TF_STATE_CONFIG_MAP_NAME
    curl \
      --silent \
      --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
      --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
      --header "Content-Type: application/yaml" \
      --request PUT \
      --data-binary @/tf-state-out/$TF_STATE_CONFIG_MAP_NAME \
      https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT_HTTPS/api/v1/namespaces/$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)/configmaps/$TF_STATE_CONFIG_MAP_NAME > /dev/null

    # validate that the current terraform state is properly reflected in the config map
    curl \
      --silent \
      --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
      --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
      --header "Accept: application/yaml" \
      --request GET https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT_HTTPS/api/v1/namespaces/$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)/configmaps/$TF_STATE_CONFIG_MAP_NAME > /tf-state-out/$TF_STATE_CONFIG_MAP_NAME.put

    sed -i -n 's/^    //gp' /tf-state-out/$TF_STATE_CONFIG_MAP_NAME.put
    if diff /tf-state-out/terraform.tfstate /tf-state-out/$TF_STATE_CONFIG_MAP_NAME.put 1> /dev/null; then # passes (returns 0) if there is no diff
      # indicate success (exit code gets lost as the surrounding command pipes this script through 'tee')
      echo -e "\nConfigMap successfully updated with terraform state."
      if [[ $exitcode -eq 0 ]]; then
        touch /success
      fi
      exit 0
    else
      # dump terraform state so that we can find it at least in the logs
      echo -e "\nConfigMap could not be updated with terraform state! Dumping state file now to have it in the logs:"
      cat /tf-state-out/terraform.tfstate
      exit 1
    fi
  fi
}

if [[ "$command" == "validate" ]]; then
  terraform \
    validate \
    -check-variables=true \
    -var-file=/tfvars/terraform.tfvars \
    /tf
  if [[ "$?" == "0" ]]; then
    terraform \
      plan \
      -parallelism=4 \
      -detailed-exitcode \
      -state=/tf-state-tmp/terraform.tfstate \
      -var-file=/tfvars/terraform.tfvars \
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
      -state=/tf-state-tmp/terraform.tfstate \
      -state-out=/tf-state-out/terraform.tfstate \
      -var-file=/tfvars/terraform.tfvars \
      /tf
    exitcode=$?
  elif [[ "$command" == "destroy" ]]; then
    terraform \
      destroy \
      -force \
      -parallelism=4 \
      -state=/tf-state-tmp/terraform.tfstate \
      -state-out=/tf-state-out/terraform.tfstate \
      -var-file=/tfvars/terraform.tfvars \
      /tf
    exitcode=$?
  else
    # Delete trap handler - nothing to do
    trap - HUP INT QUIT PIPE TERM EXIT
    echo -e "\nUnknown command: $command"
    exit 1
  fi
fi
