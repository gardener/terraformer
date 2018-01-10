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

PROJECT          := garden-terraformer
VERSION          := v$(shell head -1 Dockerfile | cut -d ':' -f 2)
REGISTRY         := eu.gcr.io/sap-cloud-platform-dev1
IMAGE_REPOSITORY := $(REGISTRY)/garden/$(PROJECT)
IMAGE_TAG        := $(VERSION)-4

.PHONY: build
build: docker-image

.PHONY: release
release: bundle docker-image docker-login docker-push bundle-clean

.PHONY: docker-image
docker-image:
	@docker build -t $(IMAGE_REPOSITORY):$(IMAGE_TAG) --rm .

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-push
docker-push:
	@if ! docker images $(IMAGE_REPOSITORY) | awk '{ print $$2 }' | grep -q -F $(IMAGE_TAG); then echo "$(IMAGE_REPOSITORY) version $(IMAGE_TAG) is not yet built. Please run 'make docker-image'"; false; fi
	@gcloud docker -- push $(IMAGE_REPOSITORY):$(IMAGE_TAG)

.PHONY: bundle
bundle:
	@./scripts/fetch-providers

.PHONY: bundle-clean
bundle-clean:
	@rm -f terraform-provider*
	@rm -f terraform
	@rm -f terraform*.zip
