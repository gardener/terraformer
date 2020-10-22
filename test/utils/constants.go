// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

// these consts are replicated from the production to not reuse them in tests
const (
	// ConfigMainKey is the key for the main.tf file
	ConfigMainKey = "main.tf"
	// ConfigVarsKey is the key for the variables.tf file
	ConfigVarsKey = "variables.tf"
	// VarsKey is the key for the terraform.tfvars file
	VarsKey = "terraform.tfvars"
	// StateKey is the key for the terraform.tfstate file
	StateKey = "terraform.tfstate"
)
