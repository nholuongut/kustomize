// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package builtinpluginconsts

const (
	imagesFieldSpecs = `
images:
- path: spec/containers[]/image
  create: true
- path: spec/initContainers[]/image
  create: true
- path: spec/template/spec/containers[]/image
  create: true
- path: spec/template/spec/initContainers[]/image
  create: true
`
)
