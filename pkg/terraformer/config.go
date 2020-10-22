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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/terraformer/pkg/utils"
)

const (
	tfConfigMainKey = "main.tf"
	tfConfigVarsKey = "variables.tf"
	tfVarsKey       = "terraform.tfvars"
	tfStateKey      = "terraform.tfstate"
)

// EnsureTFDirs ensures that the needed directories for the terraform files are present.
func (t *Terraformer) EnsureTFDirs() error {
	return t.paths.EnsureDirs(t.stepLogger("EnsureTFDirs"))
}

// FetchConfigAndState fetches the needed config and state objects from the Kubernetes API and stores their contents
// in separate files.
func (t *Terraformer) FetchConfigAndState(ctx context.Context) error {
	log := t.stepLogger("FetchConfigAndState")

	var (
		wg      wait.Group
		allErrs = &multierror.Error{
			ErrorFormat: utils.NewErrorFormatFuncWithPrefix("failed to fetch terraform config"),
		}
		errCh = make(chan error)
	)

	wg.Start(func() {
		errCh <- fetchConfigMap(ctx, log, t.client, t.config.Namespace, t.config.ConfigurationConfigMapName, false,
			t.paths.ConfigDir, tfConfigMainKey, tfConfigVarsKey,
		)
	})
	wg.Start(func() {
		errCh <- fetchConfigMap(ctx, log, t.client, t.config.Namespace, t.config.StateConfigMapName, true,
			t.paths.StateDir, tfStateKey,
		)
	})
	wg.Start(func() {
		errCh <- fetchSecret(ctx, log, t.client, t.config.Namespace, t.config.VariablesSecretName, false,
			t.paths.VarsDir, tfVarsKey,
		)
	})

	go func() {
		defer close(errCh)
		wg.Wait()
	}()

	for err := range errCh {
		allErrs = multierror.Append(allErrs, err)
	}

	return allErrs.ErrorOrNil()
}

func fetchConfigMap(ctx context.Context, log logr.Logger, c client.Client, ns, name string, optional bool, dir string, dataKeys ...string) error {
	return fetchObject(ctx, log, c, "ConfigMap", ns, name, &ConfigMapStore{&corev1.ConfigMap{}}, optional, dir, dataKeys...)
}

func fetchSecret(ctx context.Context, log logr.Logger, c client.Client, ns, name string, optional bool, dir string, dataKeys ...string) error {
	return fetchObject(ctx, log, c, "Secret", ns, name, &SecretStore{&corev1.Secret{}}, optional, dir, dataKeys...)
}

func fetchObject(ctx context.Context, log logr.Logger, c client.Client, kind, ns, name string, obj Store, optional bool, dir string, dataKeys ...string) error {
	key := client.ObjectKey{Namespace: ns, Name: name}
	log = log.WithValues("kind", kind, "object", key, "dir", dir)
	log.V(1).Info("fetching object")

	if err := c.Get(ctx, key, obj.Object()); err != nil {
		if apierrors.IsNotFound(err) && optional {
			log.V(1).Info("object not found but optional")

			// if terraformer container was restarted, the state file should always be truncated and filled with the contents
			// of the state ConfigMap (which is the single source of truth)
			for _, dataKey := range dataKeys {
				filePath := filepath.Join(dir, dataKey)
				log.V(1).Info("creating empty file / truncating existing file", "dataKey", dataKey, "file", filePath)
				file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				if err := file.Close(); err != nil {
					return err
				}
			}

			return nil
		}
		return err
	}

	for _, dataKey := range dataKeys {
		filePath := filepath.Join(dir, dataKey)

		if err := func() error {
			var notFoundErr KeyNotFoundError
			reader, err := obj.Read(dataKey)
			if errors.As(err, &notFoundErr) && optional {
				log.V(1).Info("key not found but object is optional, creating empty file / truncating existing file", "dataKey", dataKey, "file", filePath)
				file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				return file.Close()
			}
			if err != nil {
				return fmt.Errorf("failed reading from %s %q: %w", kind, key, err)
			}

			log.V(1).Info("copying contents to file", "dataKey", dataKey, "file", filePath)
			file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, reader)
			return err
		}(); err != nil {
			return err
		}
	}

	return nil
}

// StoreState stores the state file in the configured state ConfigMap.
func (t *Terraformer) StoreState(ctx context.Context) error {
	return storeConfigMap(ctx, t.log, t.client, t.config.Namespace, t.config.StateConfigMapName, t.paths.StateDir, tfStateKey)
}

func storeConfigMap(ctx context.Context, log logr.Logger, c client.Client, ns, name string, dir string, dataKeys ...string) error {
	return storeObject(ctx, log.WithValues("kind", "ConfigMap"), c,
		&ConfigMapStore{&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}}}, dir, dataKeys...)
}

func storeObject(ctx context.Context, log logr.Logger, c client.Client, obj Store, dir string, dataKeys ...string) error {
	key := client.ObjectKey{Namespace: obj.Object().GetNamespace(), Name: obj.Object().GetName()}
	log = log.WithValues("object", key)

	objBefore := obj.Object().DeepCopyObject()
	for _, dataKey := range dataKeys {
		if err := func() error {
			file, err := os.Open(filepath.Join(dir, dataKey))
			if err != nil {
				return err
			}
			defer file.Close()
			log.V(1).Info("copying file content into object", "dataKey", dataKey, "file", file.Name())

			return obj.Store(dataKey, file)
		}(); err != nil {
			return err
		}
	}

	log.V(1).Info("updating object")
	if err := func() error {
		// TODO: figure out the right way to do retries here
		if err := c.Create(ctx, obj.Object()); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return err
			}
		} else {
			return nil
		}

		if err := c.Patch(ctx, obj.Object(), client.MergeFrom(objBefore)); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}
	log.V(1).Info("successfully updated object")
	return nil
}

// StartFileWatcher watches the state file for changes and stores the file contents in the state ConfigMap as soon as
// the file gets updated.
func (t *Terraformer) StartFileWatcher(ctx context.Context, wg *sync.WaitGroup) error {
	log := t.stepLogger("fileWatcher")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.V(1).Info("stopping file watcher")
				_ = watcher.Close()
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				fileLog := log.WithValues("file", event.Name)
				fileLog.V(1).Info("received event for file", "op", event.Op.String())

				if event.Name == t.paths.StatePath && event.Op&fsnotify.Write == fsnotify.Write {
					fileLog.V(1).Info("trigger storing state")

					if err := func() error {
						// run StoreState in background, ctx might have been cancelled already
						storeCtx, storeCancel := context.WithTimeout(context.Background(), continuousStateUpdateTimeout)
						defer storeCancel()

						return t.StoreState(storeCtx)
					}(); err != nil {
						log.Error(err, "failed to store terraform state after state file update")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error(err, "error while watching state file")
			}
		}
	}()

	log.Info("starting file watcher for state file", "file", t.paths.StatePath)
	return watcher.Add(t.paths.StatePath)
}
