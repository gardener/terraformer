# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# golang-base
ARG PROVIDER=all
FROM golang:1.19.6 AS golang-base

############# terraform-bundle
FROM golang-base AS terraform-base

# install unzip (needed for unzipping terraform provider plugins)
RUN apt-get update && \
    apt-get install -y unzip

WORKDIR /tmp/terraformer
COPY ./build/fetch-providers.sh .

# overwrite to build provider-specific image
ARG PROVIDER

# copy provider-specifc TF_VERSION
COPY ./build/$PROVIDER/TF_VERSION .

############# terraform-fetch
FROM terraform-base as terraform-providers

# install terraform and needed provider plugins
RUN mkdir -p /go/src/github.com/hashicorp && \
    git clone --single-branch --depth 1 --branch v0.15.5 https://github.com/hashicorp/terraform.git /go/src/github.com/hashicorp/terraform && \
    cd /go/src/github.com/hashicorp/terraform && \
    go install ./tools/terraform-bundle

# overwrite to build provider-specific image
ARG PROVIDER

# copy provider-specific terraform-bundle.hcl
COPY ./build/$PROVIDER/terraform-bundle.hcl .
RUN ./fetch-providers.sh

############# terraform-runtime
FROM terraform-base as terraform-runtime

# overwrite to build provider-specific image
ARG PROVIDER

# copy provider-specifc TF_VERSION
COPY ./build/$PROVIDER/TF_VERSION .

RUN export TF_VERSION=$(cat ./TF_VERSION) && mkdir -p /go/src/github.com/hashicorp && \
    git clone --single-branch --depth 1 --branch v${TF_VERSION} https://github.com/hashicorp/terraform.git /go/src/github.com/hashicorp/terraform && \
    cd /go/src/github.com/hashicorp/terraform && \
    go install .

############# builder
FROM golang-base AS builder

WORKDIR /go/src/github.com/gardener/terraformer
COPY . .

ARG PROVIDER

RUN make install PROVIDER=$PROVIDER

############# terraformer
FROM alpine:3.17.2 AS terraformer

# add additional packages that are required by provider plugins
RUN apk add --update tzdata

WORKDIR /

ENV TF_DEV=true
ENV TF_RELEASE=true

COPY hack/terraform.rc /.terraform.rc
COPY --from=terraform-providers /tmp/terraformer/tfproviders/ /terraform-providers/
COPY --from=terraform-runtime /go/bin/terraform /bin/terraform
COPY --from=builder /go/bin/terraformer /

ENTRYPOINT ["/terraformer"]

############# dev
FROM golang-base AS dev

WORKDIR /go/src/github.com/gardener/terraformer
VOLUME /go/src/github.com/gardener/terraformer

COPY hack/terraform.rc ./.terraform.rc
COPY --from=terraform-providers /tmp/terraformer/tfproviders/ /terraform-providers/
COPY --from=terraform-runtime /go/bin/terraform /bin/terraform

COPY vendor vendor
COPY Makefile VERSION go.mod go.sum ./