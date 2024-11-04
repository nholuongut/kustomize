// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package types

// ConfigMapArgs contains the metadata of how to generate a configmap.
type ConfigMapArgs struct {
	// GeneratorArgs for the configmap.
	GeneratorArgs `json:",inline,omitempty" yaml:",inline,omitempty"`
}
