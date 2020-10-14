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
	"bytes"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// KeyNotFoundError is returned from a Store.Read if the store does not contain a value for the requested key.
type KeyNotFoundError string

func (k KeyNotFoundError) Error() string {
	return fmt.Sprintf("key %q not found", string(k))
}

// Store models storing arbitrary data in a runtime.Object. Implementations have to define how to read and store data
// under a certain key.
type Store interface {
	// Object returns the underlying Object to pass it to a client (for retrieving and updating).
	Object() controllerutil.Object
	// Read returns a reader for reading the contents stored under the given key.
	Read(key string) (io.Reader, error)
	// Store reads the given data and stores it under the given key.
	Store(key string, data io.Reader) error
}

var _ Store = &ConfigMapStore{}

// ConfigMapStore implements Store by storing data in a ConfigMap.
type ConfigMapStore struct {
	*corev1.ConfigMap
}

func (c *ConfigMapStore) Object() controllerutil.Object {
	return c.ConfigMap
}

func (c *ConfigMapStore) Read(key string) (io.Reader, error) {
	data, ok := c.Data[key]
	if !ok {
		return nil, KeyNotFoundError(key)
	}

	return strings.NewReader(data), nil
}

func (c *ConfigMapStore) Store(key string, data io.Reader) error {
	if c.ConfigMap.Data == nil {
		c.ConfigMap.Data = make(map[string]string, 1)
	}

	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, data)
	if err != nil {
		return err
	}

	c.ConfigMap.Data[key] = buf.String()
	return nil
}

var _ Store = &SecretStore{}

// SecretStore implements Store by storing data in a Secret.
type SecretStore struct {
	*corev1.Secret
}

func (s *SecretStore) Object() controllerutil.Object {
	return s.Secret
}

func (s *SecretStore) Read(key string) (io.Reader, error) {
	data, ok := s.Data[key]
	if !ok {
		return nil, KeyNotFoundError(key)
	}

	return bytes.NewReader(data), nil
}

func (s *SecretStore) Store(key string, data io.Reader) error {
	if s.Secret.Data == nil {
		s.Secret.Data = make(map[string][]byte, 1)
	}

	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, data)
	if err != nil {
		return err
	}

	s.Secret.Data[key] = buf.Bytes()
	return nil
}
