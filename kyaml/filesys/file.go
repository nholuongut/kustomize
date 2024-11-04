// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"io"
	"os"
)

// File groups the basic os.File methods.
type File interface {
	io.ReadWriteCloser
	Stat() (os.FileInfo, error)
}
