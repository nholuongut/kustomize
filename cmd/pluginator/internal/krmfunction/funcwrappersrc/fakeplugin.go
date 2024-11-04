// Copyright 2022 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package funcwrappersrc

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

type plugin struct{}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
	_ *resmap.PluginHelpers, _ []byte) (err error) {
	return nil
}

func (p *plugin) Transform(_ resmap.ResMap) error {
	return nil
}
