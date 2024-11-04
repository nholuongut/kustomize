// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package ext

// IgnoreFileName returns the name for ignore files in
// packages. It can be overridden by tools using this library.
var IgnoreFileName = func() string {
	return ".krmignore"
}
