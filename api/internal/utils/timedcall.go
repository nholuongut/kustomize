// Copyright 2020 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"time"
)

// TimedCall runs fn, failing if it doesn't complete in the given duration.
// The description is used in the timeout error message.
func TimedCall(description string, d time.Duration, fn func() error) error {
	done := make(chan error, 1)
	timer := time.NewTimer(d)
	defer timer.Stop()
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		return NewErrTimeOut(d, description)
	}
}
