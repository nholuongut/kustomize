// Copyright 2021 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package suffix_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/suffix"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var mutationTrackerStub = filtertest_test.MutationTrackerStub{}

var tests = map[string]TestCase{
	"suffix": {
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-foo
`,
		filter: suffix.Filter{
			Suffix:    "-foo",
			FieldSpec: types.FieldSpec{Path: "metadata/name"},
		},
	},

	"data-fieldspecs": {
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
a:
  b:
    c: d
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
a:
  b:
    c: d
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
a:
  b:
    c: d-foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
a:
  b:
    c: d-foo
`,
		filter: suffix.Filter{
			Suffix:    "-foo",
			FieldSpec: types.FieldSpec{Path: "a/b/c"},
		},
	},

	"mutation tracker": {
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-foo
`,
		filter: suffix.Filter{
			Suffix:    "-foo",
			FieldSpec: types.FieldSpec{Path: "metadata/name"},
		},
		mutationTracker: mutationTrackerStub.MutationTracker,
		expectedSetValueArgs: []filtertest_test.SetValueArg{
			{
				Value:    "instance-foo",
				NodePath: []string{"metadata", "name"},
			},
			{
				Value:    "instance-foo",
				NodePath: []string{"metadata", "name"},
			},
		},
	},
}

type TestCase struct {
	input                string
	expected             string
	filter               suffix.Filter
	mutationTracker      func(key, value, tag string, node *yaml.RNode)
	expectedSetValueArgs []filtertest_test.SetValueArg
}

func TestFilter(t *testing.T) {
	for name := range tests {
		mutationTrackerStub.Reset()
		test := tests[name]
		test.filter.WithMutationTracker(test.mutationTracker)
		t.Run(name, func(t *testing.T) {
			if !assert.Equal(t,
				strings.TrimSpace(test.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, test.input, test.filter))) {
				t.FailNow()
			}
			assert.Equal(t, test.expectedSetValueArgs, mutationTrackerStub.SetValueArgs())
		})
	}
}
