// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package loader_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/loader"
	. "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	//nolint:gosec
	secretGenerator = `
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: secretGenerator
name: mySecret
behavior: merge
envFiles:
- a.env
- b.env
valueFiles:
- longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
`
	someServiceGenerator = `
apiVersion: someteam.example.com/v1
kind: SomeServiceGenerator
metadata:
  name: myServiceGenerator
service: my-service
port: "12345"
`
)

func TestLoader(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("builtin", "", "SecretGenerator").
		BuildGoPlugin("someteam.example.com", "v1", "SomeServiceGenerator")
	defer th.Reset()
	p := provider.NewDefaultDepProvider()
	rmF := resmap.NewFactory(p.GetResourceFactory())
	fsys := filesys.MakeFsInMemory()
	fLdr, err := loader.NewLoader(
		loader.RestrictionRootOnly,
		filesys.Separator, fsys)
	if err != nil {
		t.Fatal(err)
	}
	generatorConfigs, err := rmF.NewResMapFromBytes([]byte(
		someServiceGenerator + "---\n" + secretGenerator))
	if err != nil {
		t.Fatal(err)
	}
	for _, behavior := range []types.BuiltinPluginLoadingOptions{
		/* types.BploUseStaticallyLinked,
		types.BploLoadFromFileSys */} {
		c := types.EnabledPluginConfig(behavior)
		pLdr := NewLoader(c, rmF, fsys)
		if pLdr == nil {
			t.Fatal("expect non-nil loader")
		}
		_, err = pLdr.LoadGenerators(
			fLdr, valtest_test.MakeFakeValidator(), generatorConfigs)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoaderWithWorkingDir(t *testing.T) {
	p := provider.NewDefaultDepProvider()
	rmF := resmap.NewFactory(p.GetResourceFactory())
	fsys := filesys.MakeFsInMemory()
	c := types.EnabledPluginConfig(types.BploLoadFromFileSys)
	pLdr := NewLoader(c, rmF, fsys)
	npLdr := pLdr.LoaderWithWorkingDir("/tmp/dummy")
	require.Equal(t,
		"",
		pLdr.Config().FnpLoadingOptions.WorkingDir,
		"the plugin working dir should not change")
	require.Equal(t,
		"/tmp/dummy",
		npLdr.Config().FnpLoadingOptions.WorkingDir,
		"the plugin working dir is not updated")
}
