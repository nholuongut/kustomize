# Copyright 2019 Nho Luong DevOps.
# SPDX-License-Identifier: Apache-2.0

include ../Makefile-examples-base.mk

.PHONY: image
image:
	docker build image -t gcr.io/kustomize-functions/example-nginx:v0.2.0
	docker push gcr.io/kustomize-functions/example-nginx:v0.2.0
