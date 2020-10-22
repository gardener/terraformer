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

package terraformer

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Command is a terraform command
type Command string

// known terraform commands
const (
	Init     Command = "init"
	Apply    Command = "apply"
	Destroy  Command = "destroy"
	Validate Command = "validate" // TODO: is this still needed?
	Plan     Command = "plan"
)

// supported terraform commands, that can be run as `terraformer <command>`
var SupportedCommands = map[Command]struct{}{
	Apply:    {},
	Destroy:  {},
	Validate: {},
}

// Terraformer can execute terraform commands and fetch/store config and state from/into Secrets/ConfigMaps
type Terraformer struct {
	config *Config
	paths  *PathSet
	log    logr.Logger

	client client.Client
}

type Config struct {
	// ConfigurationConfigMapName is the name of the ConfigMap that holds the `main.tf` and `variables.tf` files.
	ConfigurationConfigMapName string
	// StateConfigMapName is the name of the ConfigMap that the `terraform.tfstate` file should be stored in.
	StateConfigMapName string
	// VariablesSecretName is the name of the Secret that holds the `terraform.tfvars` file.
	VariablesSecretName string
	// Namespace is the namespace to store the configuration resources in.
	Namespace string

	// RESTConfig holds the completed rest.Config.
	RESTConfig *rest.Config
}

// MarshalLogObject implements zapcore.ObjectMarshaler
func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("configurationConfigMapName", c.ConfigurationConfigMapName)
	enc.AddString("stateConfigMapName", c.StateConfigMapName)
	enc.AddString("variablesSecretName", c.VariablesSecretName)
	enc.AddString("namespace", c.Namespace)
	return nil
}