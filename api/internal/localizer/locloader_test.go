// Copyright 2022 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	. "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func checkNewLoader(req *require.Assertions, ldr *Loader, args *Args, target string, scope string, newDir string, fSys filesys.FileSystem) {
	checkLoader(req, ldr, target)
	checkArgs(req, args, target, scope, newDir, fSys)
}

func checkLoader(req *require.Assertions, ldr ifc.Loader, root string) {
	req.Equal(root, ldr.Root())
	req.Empty(ldr.Repo())
}

func checkArgs(req *require.Assertions, args *Args, target string, scope string, newDir string, fSys filesys.FileSystem) {
	req.Equal(target, args.Target.String())
	req.Equal(scope, args.Scope.String())
	req.Equal(newDir, args.NewDir.String())
	req.True(fSys.Exists(newDir))
}

func TestLocalLoadNewAndCleanup(t *testing.T) {
	req := require.New(t)
	fSys := makeMemoryFs(t)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	// typical setup
	ldr, args, err := NewLoader("a", "/", "/newDir", fSys)
	req.NoError(err)
	checkNewLoader(req, ldr, &args, "/a", "/", "/newDir", fSys)

	fSysCopy := makeMemoryFs(t)
	req.NoError(fSysCopy.Mkdir("/newDir"))
	req.Equal(fSysCopy, fSys)

	// easy load directly in root
	content, err := ldr.Load("pod.yaml")
	req.NoError(err)
	req.Equal([]byte(podConfiguration), content)

	// typical sibling root reference
	sibLdr, err := ldr.New("../alpha")
	req.NoError(err)
	checkLoader(req, sibLdr, "/alpha")

	// only need to test once, since don't need to call Cleanup() on local target
	req.NoError(sibLdr.Cleanup())
	req.NoError(ldr.Cleanup())

	// file system and buffer checks are also one-time
	req.Equal(fSysCopy, fSys)
	req.Empty(buf.String())
}

func TestNewLocLoaderDefaultForRootTarget(t *testing.T) {
	cases := map[string]struct {
		target string
		scope  string
	}{
		"explicit": {
			"/",
			".",
		},
		"implicit": {
			".",
			"",
		},
	}
	for name, params := range cases {
		params := params
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			fSys := makeMemoryFs(t)

			ldr, args, err := NewLoader(params.target, params.scope, "", fSys)
			req.NoError(err)
			checkNewLoader(req, ldr, &args, "/", "/", "/"+DstPrefix, fSys)

			// file in root, but nested
			content, err := ldr.Load("a/pod.yaml")
			req.NoError(err)
			req.Equal([]byte(podConfiguration), content)

			childLdr, err := ldr.New("a")
			req.NoError(err)
			checkLoader(req, childLdr, "/a")

			// messy, uncleaned path
			content, err = childLdr.Load("./../a/pod.yaml")
			req.NoError(err)
			req.Equal([]byte(podConfiguration), content)
		})
	}
}

func TestNewMultiple(t *testing.T) {
	req := require.New(t)
	fSys := makeMemoryFs(t)

	// default destination for non-file system root target
	// destination outside of scope
	ldr, args, err := NewLoader("/alpha/beta", "/alpha", "", fSys)
	req.NoError(err)
	checkNewLoader(req, ldr, &args, "/alpha/beta", "/alpha", "/"+DstPrefix+"-beta", fSys)

	// nested child root that isn't cleaned
	descLdr, err := ldr.New("../beta/gamma/delta")
	req.NoError(err)
	checkLoader(req, descLdr, "/alpha/beta/gamma/delta")

	// upwards traversal
	higherLdr, err := descLdr.New("../../say")
	req.NoError(err)
	checkLoader(req, higherLdr, "/alpha/beta/say")
}

func makeWdFs(t *testing.T) map[string]filesys.FileSystem {
	t.Helper()
	req := require.New(t)

	root := filesys.MakeEmptyDirInMemory()
	req.NoError(root.MkdirAll("a/b/c/d/e"))

	outer, err := root.Find("a")
	req.NoError(err)
	middle, err := root.Find("a/b/c")
	req.NoError(err)

	return map[string]filesys.FileSystem{
		"a":     outer,
		"a/b/c": middle,
	}
}

func TestNewLocLoaderCwdNotRoot(t *testing.T) {
	cases := map[string]struct {
		wd     string
		target string
		scope  string
		newDir string
	}{
		// target not immediate child of scope
		"outer dir": {
			"a",
			"b/c/d/e",
			"b/c",
			"b/newDir",
		},
		"scope": {
			"a/b/c",
			"d/e",
			".",
			"d/e/newDir",
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			fSys := makeWdFs(t)[test.wd]

			ldr, args, err := NewLoader(test.target, test.scope, test.newDir, fSys)
			req.NoError(err)
			checkLoader(req, ldr, "a/b/c/d/e")

			req.Equal("a/b/c/d/e", args.Target.String())
			req.Equal("a/b/c", args.Scope.String())
			req.Equal(test.wd+"/"+test.newDir, args.NewDir.String())
			// memory file system can only find paths rooted at current node
			req.True(fSys.Exists(test.newDir))
		})
	}
}

func TestNewLocLoaderFails(t *testing.T) {
	cases := map[string]struct {
		target string
		scope  string
		dest   string
	}{
		"non-existent target": {
			"/b",
			"/",
			"/newDir",
		},
		"file target": {
			"/a/pod.yaml",
			"/",
			"/newDir",
		},
		"inner scope": {
			"/alpha",
			"/alpha/beta",
			"/newDir",
		},
		"side scope": {
			"/alpha",
			"/a",
			"/newDir",
		},
		"existing dst": {
			"/alpha",
			"/",
			"/a",
		},
	}
	for name, params := range cases {
		params := params
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer func() {
				log.SetOutput(os.Stderr)
			}()

			_, _, err := NewLoader(params.target, params.scope, params.dest, makeMemoryFs(t))
			require.Error(t, err)
			require.Empty(t, buf.String())
		})
	}
}

func TestNewFails(t *testing.T) {
	req := require.New(t)
	fSys := makeMemoryFs(t)

	ldr, args, err := NewLoader("/alpha/beta/gamma", "alpha", "alpha/beta/gamma/newDir", fSys)
	req.NoError(err)
	checkNewLoader(req, ldr, &args, "/alpha/beta/gamma", "/alpha", "/alpha/beta/gamma/newDir", fSys)

	cases := map[string]string{
		"outside scope":     "../../../a",
		"at dst":            "newDir",
		"ancestor":          "../../beta",
		"non-existent root": "delt",
		"file":              "delta/deployment.yaml",
	}
	for name, root := range cases {
		root := root
		t.Run(name, func(t *testing.T) {
			fSys := makeMemoryFs(t)

			ldr, _, err := NewLoader("/alpha/beta/gamma", "alpha", "alpha/beta/gamma/newDir", fSys)
			require.NoError(t, err)

			_, err = ldr.New(root)
			require.Error(t, err)
		})
	}
}

func TestLoadFails(t *testing.T) {
	req := require.New(t)
	fSys := makeMemoryFs(t)

	ldr, args, err := NewLoader("./a/../a", "/a/../a", "/a/newDir", fSys)
	req.NoError(err)
	checkNewLoader(req, ldr, &args, "/a", "/a", "/a/newDir", fSys)

	cases := map[string]string{
		"directory":          "b",
		"non-existent file":  "kubectl.yaml",
		"file outside root":  "../alpha/beta/gamma/delta/deployment.yaml",
		"inside dst":         "newDir/pod.yaml",
		"winding inside dst": "/a/test/../newDir/pod.yaml",
	}
	for name, file := range cases {
		file := file
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			fSys := makeMemoryFs(t)

			ldr, _, err := NewLoader("./a/../a", "/a/../a", "/a/newDir", fSys)
			req.NoError(err)

			_, err = ldr.Load(file)
			req.Error(err)
		})
	}
}
