# Terraformer

The Terraformer is a tool that can execute Terraform configuration and is designed to run as a pod inside a Kubernetes cluster. The `main.tf`, `variables.tf`, `terraform.tfvars` and `terraform.tfstate` files are expected to be stored as configmaps/secrets and mounted into the pod. The Terraformer is able to run `terraform validate|plan|apply|destroy` and will update the state (configmap) itself by using the available Kubernetes service account.

Usually, one will run `terraform validate|plan` within a single pod and `terraform apply|destroy` as a job in order to establish retry logic.

## Constraints

The `main.tf` and `variables.tf` files are expected inside the pod at `/tf` location, the `terraform.tfvars` is expected at `/tfvars` and the `terraform.tfstate` is expected at `/tf-state-in`.

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

This will bundle all Terraform provider plugins, create a new Docker image with the tag you specified in the `Makefile`, push it to our image registry, and clean up afterwards.

## Example manifests

Please find example Kubernetes manifests within the [`/example`](example) directory.
