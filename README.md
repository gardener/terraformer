# Terraformer

The Terraformer is a tool that can execute Terraform configuration and is designed to run as a pod inside a Kubernetes cluster. The `main.tf`, `variables.tf`, `terraform.tfvars` and `terraform.tfstate` files are expected to be stored as configmaps/secrets and mounted into the pod. The Terraformer is able to run `terraform validate|plan|apply|destroy` and will update the state (configmap) itself by using the available Kubernetes service account.

Usually, `terraform validate|plan` and `terraform apply|destroy` run within a single Pod. Note that running `terraform apply|destroy` as a Job is not recommended. The Job object will start a new Pod if the first Pod fails or is deleted (for example due to a node hardware failure or a node reboot). You may end in a case with two Terraformer Pods running at the same time which can fail with conflicts.

## Constraints

The `main.tf` and `variables.tf` files are expected inside the pod at `/tf` location, the `terraform.tfvars` is expected at `/tfvars` and the `terraform.tfstate` is expected at `/tf-state-in`.

## Example manifests

Example Kubernetes manifests can be found within the [`/example`](example) directory.

## Images

Terraformer images are pushed to gcr.io. The list of existing images and version can be found in [eu.gcr.io/gardener-project/gardener/terraformer](https://eu.gcr.io/gardener-project/gardener/terraformer).

## How to build it?

It is required that you have installed the `terraform-bundle` binary. You can get it from the [offical Hashicorp Terraform](https://github.com/hashicorp/terraform) repository:

```bash
$ go get github.com/hashicorp/terraform
$ cd $GOPATH/src/github.com/hashicorp/terraform
$ go install ./tools/terraform-bundle
```

:warning: Please don't forget to update the `VERSION` file before creating a new release:

```bash
$ make release
```

This will bundle all Terraform provider plugins, create a new Docker image with the tag you specified in the `Makefile`, push it to the specified image registry, and clean up afterwards.
