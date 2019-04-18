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

terraform {
  version = "TF_VERSION"
}

providers {
  aws       = ["1.60.0"]
  azurerm   = ["1.22.1"]
  google    = ["1.20.0"]
  openstack = ["1.16.0"]
  alicloud  = ["1.31.0"]
  packet    = ["1.7.2"]
  template  = ["1.0.0"]
  null      = ["1.0.0"]
}
