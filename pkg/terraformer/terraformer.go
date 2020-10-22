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
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/terraformer/pkg/utils"
)

var (
	// TerraformBinary is the name of the terraform binary, it allows to overwrite it for testing purposes
	TerraformBinary = "terraform"

	// allow redirecting output in tests to GinkgoWriter
	Stdout, Stderr io.Writer = os.Stdout, os.Stderr
)

const (
	continuousStateUpdateTimeout = 5 * time.Minute
	finalStateUpdateTimeout      = 5 * time.Minute
)

// NewDefaultTerraformer creates a new Terraformer with the default PathSet and logger.
func NewDefaultTerraformer(config *Config) (*Terraformer, error) {
	return NewTerraformer(config, runtimelog.Log, DefaultPaths())
}

// NewTerraformer creates a new Terraformer with the given options.
func NewTerraformer(config *Config, log logr.Logger, paths *PathSet) (*Terraformer, error) {
	t := &Terraformer{
		config: config,
		log:    log,
		paths:  paths,
	}

	c, err := client.New(config.RESTConfig, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	t.client = c

	return t, nil
}

func (t *Terraformer) Run(command Command) error {
	if _, ok := SupportedCommands[command]; !ok {
		return fmt.Errorf("terraform command %q is not supported", command)
	}

	t.log.V(1).Info("executing terraformer with config", "config", t.config)

	return t.execute(command)
}

func (t *Terraformer) execute(command Command) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-sigCh:
			t.log.Info("interrupt received")
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := t.EnsureTFDirs(); err != nil {
		return fmt.Errorf("failed to create needed directories: %w", err)
	}

	if err := t.FetchConfigAndState(ctx); err != nil {
		return err
	}

	var fileWatcherWaitGroup sync.WaitGroup
	// file watcher should run in background and should only be cancelled when this function returns, i.e. when any
	// running terraform processes have finished
	fileWatcherCtx, fileWatcherCancel := context.WithCancel(context.Background())

	// always try to store state once again before exiting
	defer func() {
		// stop file watcher and wait for it to be done before storing state explicitly to avoid conflicts
		fileWatcherCancel()
		fileWatcherWaitGroup.Wait()

		log := t.stepLogger("StoreState")
		log.Info("storing state before exiting")

		// run StoreState in background, ctx might have been cancelled already
		storeCtx, storeCancel := context.WithTimeout(context.Background(), finalStateUpdateTimeout)
		defer storeCancel()

		if err := t.StoreState(storeCtx); err != nil {
			log.Error(err, "failed to store terraform state")
		}
		log.Info("successfully stored terraform state")
	}()

	// continuously update state configmap as soon as state file changes on disk
	if err := t.StartFileWatcher(fileWatcherCtx, &fileWatcherWaitGroup); err != nil {
		return fmt.Errorf("failed to start state file watcher: %w", err)
	}

	// initialize terraform plugins
	if err := t.executeTerraform(ctx, Init); err != nil {
		return fmt.Errorf("error executing terraform %s: %w", command, err)
	}

	// execute main terraform command
	if err := t.executeTerraform(ctx, command); err != nil {
		return fmt.Errorf("error executing terraform %s: %w", command, err)
	}

	if command == Validate {
		if err := t.executeTerraform(ctx, Plan); err != nil {
			return fmt.Errorf("error executing terraform %s: %w", Plan, err)
		}
	}

	return nil
}

func (t *Terraformer) executeTerraform(ctx context.Context, command Command) error {
	log := t.stepLogger("executeTerraform")

	args := []string{string(command)}

	switch command {
	case Init:
		args = append(args, "-plugin-dir="+t.paths.ProvidersDir)
	case Plan:
		args = append(args, "-var-file="+t.paths.VarsPath, "-parallelism=4", "-detailed-exitcode", "-state="+t.paths.StatePath)
	case Apply:
		args = append(args, "-var-file="+t.paths.VarsPath, "-parallelism=4", "-auto-approve", "-state="+t.paths.StatePath)
	case Destroy:
		args = append(args, "-var-file="+t.paths.VarsPath, "-parallelism=4", "-auto-approve", "-state="+t.paths.StatePath)
	}

	args = append(args, t.paths.ConfigDir)

	log.Info("executing terraform", "command", command, "args", strings.Join(args[1:], " "))
	tfCmd := exec.Command(TerraformBinary, args...)
	tfCmd.Stdout = Stdout
	tfCmd.Stderr = Stderr

	if err := tfCmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	// wait for signal handler goroutine to finish properly before returning
	defer wg.Wait()

	doneCh := make(chan struct{})
	defer close(doneCh)

	// setup signal handler relaying signals to terraform process
	go func() {
		defer wg.Done()
		select {
		case <-doneCh:
			return
		case <-ctx.Done():
			log.V(1).Info("relaying interrupt to terraform process")
			if err := tfCmd.Process.Signal(syscall.SIGINT); err != nil {
				log.Error(err, "failed to relay interrupt to terraform process")
			}
		}
	}()

	if err := tfCmd.Wait(); err != nil {
		log.Error(err, "terraform process finished with error", "command", command)
		return utils.WithExitCode{Code: tfCmd.ProcessState.ExitCode(), Underlying: err}
	}

	log.Info("terraform process finished successfully", "command", command)
	return nil
}
