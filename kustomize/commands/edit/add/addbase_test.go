// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	baseDirectoryPaths = "my/path/to/wonderful/base,other/path/to/even/more/wonderful/base"
)

func TestAddBaseHappyPath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		err := fSys.Mkdir(base)
		require.NoError(t, err)
	}
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddBase(fSys)
	args := []string{baseDirectoryPaths}
	require.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	for _, base := range bases {
		assert.Contains(t, string(content), base)
	}
}

func TestAddBaseAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	// Create fake directories
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		err := fSys.Mkdir(base)
		require.NoError(t, err)
	}
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddBase(fSys)
	args := []string{baseDirectoryPaths}
	require.NoError(t, cmd.RunE(cmd, args))
	// adding an existing base should return an error
	require.Error(t, cmd.RunE(cmd, args))
	var expectedErrors []string
	for _, base := range bases {
		msg := "base " + base + " already in kustomization file"
		expectedErrors = append(expectedErrors, msg)
		assert.True(t, slices.Contains(expectedErrors, msg))
	}
}

func TestAddBaseNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddBase(fSys)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, "must specify a base directory", err.Error())
}
