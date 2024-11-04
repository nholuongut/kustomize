// Copyright 2020 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	sigsK8sIo = "sigs.k8s.io"

	// Files whose names start with prefixBad get special treatment from
	// https://github.com/kubernetes/kubernetes/blob/master/build/common.sh
	// (and ./hack/verify-generated-files-remake.sh, etc.).
	// We don't want that, so modify those file names.
	prefixBad  = "zz_generated"
	prefixGood = "copied"
)

type Copier struct {
	spec       *ModuleSpec
	goModCache string
	topPackage string
	pwdPackage string
	srcDir     string
	pgmName    string
}

func (c Copier) replacementPath() string {
	return filepath.Join(c.topPackage, c.subPath())
}

func (c Copier) subPath() string {
	return filepath.Join("internal", c.pwdPackage)
}

func (c Copier) Print() {
	fmt.Printf("   apiMachineryModule: %s\n", c.spec.Module)
	fmt.Printf("      replacementPath: %s\n", c.replacementPath())
	fmt.Printf("           goModCache: %s\n", c.goModCache)
	fmt.Printf("           topPackage: %s\n", c.topPackage)
	fmt.Printf("           pwdPackage: %s\n", c.pwdPackage)
	fmt.Printf("              subPath: %s\n", c.subPath())
	fmt.Printf("               srcDir: %s\n", c.srcDir)
	fmt.Printf("  apiMachineryModSpec: %s\n", c.spec.Name())
	fmt.Printf("              pgmName: %s\n", c.pgmName)
	fmt.Printf("                  pwd: %s\n", os.Getenv("PWD"))
}

func NewCopier(s *ModuleSpec, prefix, pgmName string) Copier {
	tmp := Copier{
		spec:       s,
		pgmName:    pgmName,
		pwdPackage: os.Getenv("GOPACKAGE"),
		goModCache: RunGetOutputCommand("go", "env", "GOMODCACHE"),
	}
	goMod := RunGetOutputCommand("go", "env", "GOMOD")
	topPackage := filepath.Join(goMod[:len(goMod)-len("go.mod")-1], prefix)
	k := strings.Index(topPackage, sigsK8sIo)
	if k < 1 {
		log.Fatalf("cannot find %s in %s", sigsK8sIo, topPackage)
	}
	tmp.srcDir = topPackage[:k-1]
	tmp.topPackage = topPackage[k:]
	return tmp
}

func (c Copier) CopyFile(dir, fName string) error {
	inFile, err := os.Open(
		filepath.Join(c.goModCache, c.spec.Name(), dir, fName))
	if err != nil {
		return err
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)

	newName := fName
	if strings.HasPrefix(fName, prefixBad) {
		newName = prefixGood + fName[len(prefixBad):]
	}
	w, err := newWriter(dir, newName)
	if err != nil {
		return err
	}
	defer w.close()

	w.write(
		fmt.Sprintf(
			// This particular phrasing is required.
			"// Code generated by %s from %s; DO NOT EDIT.",
			c.pgmName, c.spec.Name()))
	w.write(
		fmt.Sprintf(
			"// File content copied from %s\n",
			filepath.Join(c.spec.Name(), dir, fName)))

	for scanner.Scan() {
		l := scanner.Text()
		// Disallow recursive generation.
		if strings.HasPrefix(l, "//go:generate") ||
			strings.HasPrefix(l, "// +k8s:") {
			continue
		}
		// When copying generated code, drop the old 'generated' message.
		if strings.HasPrefix(l, "// Code generated") {
			continue
		}
		// Fix self-imports.
		l = strings.Replace(l, c.spec.Module, c.replacementPath(), 1)

		// Replace k8s.io/klog with Go's log (we must avoid k8s.io entirely).
		l = strings.Replace(l, "\"k8s.io/klog/v2\"", "\"log\"", 1)
		l = strings.Replace(l, "\"k8s.io/klog\"", "\"log\"", 1)
		l = strings.Replace(l, "klog.V(10).Infof(", "log.Printf(", 1)
		w.write(l)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	w.write("")
	return nil
}