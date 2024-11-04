// Copyright 2022 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package loader

import "sigs.k8s.io/kustomize/kyaml/errors"

var (
	ErrHTTP     = errors.Errorf("HTTP Error")
	ErrRtNotDir = errors.Errorf("must build at directory")
)
