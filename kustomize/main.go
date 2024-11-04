// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

// The kustomize CLI.
package main

import (
	"os"

	"sigs.k8s.io/kustomize/kustomize/v5/commands"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
