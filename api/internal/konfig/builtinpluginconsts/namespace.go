// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package builtinpluginconsts

const (
	namespaceFieldSpecs = `
namespace:
- path: metadata/name
  kind: Namespace
  create: true
- path: spec/service/namespace
  group: apiregistration.k8s.io
  kind: APIService
  create: true
- path: spec/conversion/webhook/clientConfig/service/namespace
  group: apiextensions.k8s.io
  kind: CustomResourceDefinition
`
)
