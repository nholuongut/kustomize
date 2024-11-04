// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package builtinpluginconsts

const replicasFieldSpecs = `
replicas:
- path: spec/replicas
  create: true
  kind: Deployment

- path: spec/replicas
  create: true
  kind: ReplicationController

- path: spec/replicas
  create: true
  kind: ReplicaSet

- path: spec/replicas
  create: true
  kind: StatefulSet
`
