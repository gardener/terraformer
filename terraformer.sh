#!/bin/bash -u
#
# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved.
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

function end_execution() {
  # Delete trap handler to avoid recursion
  trap - HUP QUIT PIPE INT TERM

  if [ -n "$PID" ] && kill -0 "$PID" &>/dev/null; then
    echo "$(date) Sending SIGTERM to terraform.sh process $PID."
    kill -SIGTERM $PID
    echo "$(date) Waiting for terraform.sh process $PID to complete..."
    wait $PID
    echo "$(date) terraform.sh process $PID completed."
  fi
}

# determine command
command="${1:-apply}"

trap end_execution INT TERM

PID=

/terraform.sh "$command" 2>&1 &
PID=$!
wait $PID
exitcode=$?

echo "$(date) terraform.sh exited with $exitcode."

[[ -f /success ]] && exit 0 || exit 1
