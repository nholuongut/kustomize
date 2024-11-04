# Copyright 2019 Nho Luong DevOps.
# SPDX-License-Identifier: Apache-2.0

include ../../../../../../Makefile-modules.mk

.PHONY: image
image:
	docker build . -t gcr.io/kustomize-functions/e2econtainerenvgenerator
	docker push gcr.io/kustomize-functions/e2econtainerenvgenerator
