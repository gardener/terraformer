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
	"os"
	"path"
	"path/filepath"

	"github.com/go-logr/logr"
)

const (
	// TODO: still needed?
	tfPluginsDir = ".terraform/plugins/linux_amd64"
)

// PathSet carries the set of file paths for terraform files and allows to set different paths in tests
type PathSet struct {
	// ConfigDir is the directory to hold the main terraform scripts (`main.tf` and `variables.tf`)
	ConfigDir string
	// VarsDir is the directory to hold the terraform variables values file (`terraform.tfvars`)
	VarsDir string
	// StateDir is the directory to hold the terraform state file (`terraform.tfstate`)
	StateDir string
	// ProvidersDir is the directory which contains the provider plugin binaries
	ProvidersDir string

	// VarsPath is the complete path the the variables values file
	VarsPath string
	// StatePath is the complete path the the state file
	StatePath string
}

// DefaultPaths returns the default PathSet used in terraformer
func DefaultPaths() *PathSet {
	p := &PathSet{
		ConfigDir:    "/tf",
		VarsDir:      "/tfvars",
		StateDir:     "/tfstate",
		ProvidersDir: "/terraform-providers",
	}
	p.VarsPath = path.Join(p.VarsDir, tfVarsKey)
	p.StatePath = path.Join(p.StateDir, tfStateKey)

	return p
}

// WithBaseDir returns a copy of the PathSet with all paths rooted in baseDir.
// This is used for testing purposes to use paths located e.g. under temporary directories.
func (p *PathSet) WithBaseDir(baseDir string) *PathSet {
	return &PathSet{
		ConfigDir:    filepath.Join(baseDir, p.ConfigDir),
		VarsDir:      filepath.Join(baseDir, p.VarsDir),
		StateDir:     filepath.Join(baseDir, p.StateDir),
		ProvidersDir: filepath.Join(baseDir, p.ProvidersDir),
		VarsPath:     filepath.Join(baseDir, p.VarsPath),
		StatePath:    filepath.Join(baseDir, p.StatePath),
	}
}

// EnsureDirs ensures that the needed directories for the terraform files are present.
func (p *PathSet) EnsureDirs(log logr.Logger) error {
	log.Info("ensuring terraform directories")

	for _, dir := range []string{
		p.ConfigDir,
		p.VarsDir,
		p.StateDir,
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		log.V(1).Info("directory ensured", "dir", dir)
	}
	return nil
}
