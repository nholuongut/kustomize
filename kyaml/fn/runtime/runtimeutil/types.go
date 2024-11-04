// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package runtimeutil

type DeferFailureFunction interface {
	GetExit() error
}
