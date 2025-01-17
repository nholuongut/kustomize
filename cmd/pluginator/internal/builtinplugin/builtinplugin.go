// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package builtinplugin

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/provenance"
)

//go:generate stringer -type=pluginType
type pluginType int

const packageForGeneratedCode = "builtins"

const (
	unknown pluginType = iota
	Transformer
	Generator
)

// ConvertToBuiltInPlugin converts the input plugin file to
// kustomize builtin plugin and writes it to proper directory
func ConvertToBuiltInPlugin() (retErr error) {
	root, err := inputFileRoot()
	if err != nil {
		return err
	}
	file, err := os.Open(root + ".go")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	err = readToPackageMain(scanner, file.Name())
	if err != nil {
		return err
	}

	w, err := newWriter(root)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := w.Close()
		if retErr == nil {
			retErr = closeErr
		}
	}()

	// This particular phrasing is required.
	w.write(
		fmt.Sprintf(
			"// Code generated by pluginator on %s; DO NOT EDIT.",
			root))
	w.write(
		fmt.Sprintf(
			"// pluginator %s\n", provenance.GetProvenance().Short()))
	w.write("package " + packageForGeneratedCode)

	pType := unknown

	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "//go:generate") {
			continue
		}
		if strings.HasPrefix(l, "var "+konfig.PluginSymbol+" plugin") {
			// Hack to skip leading new line
			scanner.Scan()
			continue
		}
		if strings.Contains(l, " Transform(") {
			if pType != unknown {
				return fmt.Errorf("unexpected Transform(")
			}
			pType = Transformer
		} else if strings.Contains(l, " Generate(") {
			if pType != unknown {
				return fmt.Errorf("unexpected Generate(")
			}
			pType = Generator
		}
		w.write(l)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	w.write("")
	w.write("func New" + root + "Plugin() resmap." + pType.String() + "Plugin {")
	w.write("	return &" + root + "Plugin{}")
	w.write("}")

	return nil
}

func inputFileRoot() (string, error) {
	n := os.Getenv("GOFILE")
	if !strings.HasSuffix(n, ".go") {
		return "", fmt.Errorf("%+v, expecting .go suffix on %s", provenance.GetProvenance(), n)
	}
	return n[:len(n)-len(".go")], nil
}

func readToPackageMain(s *bufio.Scanner, f string) error {
	gotMain := false
	for !gotMain && s.Scan() {
		gotMain = strings.HasPrefix(s.Text(), "package main")
	}
	if !gotMain {
		return fmt.Errorf("%s missing package main", f)
	}
	return nil
}

type writer struct {
	root string
	f    *os.File
}

func newWriter(r string) (*writer, error) {
	n := makeOutputFileName(r)
	f, err := os.Create(n)
	if err != nil {
		return nil, fmt.Errorf("unable to create `%s`; %v", n, err)
	}
	return &writer{root: r, f: f}, nil
}

// Assume that this command is running with a $PWD of
//
//	$HOME/kustomize/plugin/builtin/secretGenerator
//
// (for example).  Then we want to write to
//
//	$HOME/kustomize/api/builtins
func makeOutputFileName(root string) string {
	return filepath.Join(
		"..", "..", "..", "api/internal", packageForGeneratedCode, root+".go")
}

func (w *writer) Close() error {
	// Do this for debugging.
	// fmt.Println("Generated " + makeOutputFileName(w.root))
	return w.f.Close()
}

func (w *writer) write(line string) {
	_, err := w.f.WriteString(w.filter(line) + "\n")
	if err != nil {
		fmt.Printf("Trouble writing: %s", line)
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
}

func (w *writer) filter(in string) string {
	if ok, newer := w.replace(in, "type plugin struct"); ok {
		return newer
	}
	if ok, newer := w.replace(in, "*plugin)"); ok {
		return newer
	}
	return in
}

// replace 'plugin' with 'FooPlugin' in context
// sensitive manner.
func (w *writer) replace(in, target string) (bool, string) {
	if !strings.Contains(in, target) {
		return false, ""
	}
	newer := strings.Replace(
		target, "plugin", w.root+"Plugin", 1)
	return true, strings.Replace(in, target, newer, 1)
}
