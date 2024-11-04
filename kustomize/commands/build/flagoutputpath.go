// Copyright 2021 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFlagOutputPath(set *pflag.FlagSet) {
	set.StringVarP(
		&theFlags.outputPath,
		"output",
		"o", // abbreviation
		"",  // default
		"If specified, write output to this path.")
}
