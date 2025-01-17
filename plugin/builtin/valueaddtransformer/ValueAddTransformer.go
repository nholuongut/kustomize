// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/namespace"
	"sigs.k8s.io/kustomize/api/filters/valueadd"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// An 'Add' transformer inspired by the IETF RFC 6902 JSON spec Add operation.
type plugin struct {
	// Value is the value to add.
	// Defaults to base name of encompassing kustomization root.
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	// Targets is a slice of targets that should have the value added.
	Targets []Target `json:"targets,omitempty" yaml:"targets,omitempty"`

	// TargetFilePath is a file path.  If specified, the file will be parsed into
	// a slice of Target, and appended to anything that was specified in the
	// Targets field.  This is just a means to share common target specifications.
	TargetFilePath string `json:"targetFilePath,omitempty" yaml:"targetFilePath,omitempty"`
}

// Target describes where to put the value.
type Target struct {
	// Selector selects the resources to modify.
	Selector *types.Selector `json:"selector,omitempty" yaml:"selector,omitempty"`

	// NotSelector selects the resources to exclude
	// from those included by overly broad selectors.
	// TODO: implement this?
	// NotSelector *types.Selector `json:"notSelector,omitempty" yaml:"notSelector,omitempty"`

	// FieldPath is a JSON-style path to the field intended to hold the value.
	FieldPath string `json:"fieldPath,omitempty" yaml:"fieldPath,omitempty"`

	// FilePathPosition is passed to the filter directly.  Look there for doc.
	FilePathPosition int `json:"filePathPosition,omitempty" yaml:"filePathPosition,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	err := yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	p.Value = strings.TrimSpace(p.Value)
	if p.Value == "" {
		p.Value = filepath.Base(h.Loader().Root())
	}
	if p.TargetFilePath != "" {
		bytes, err := h.Loader().Load(p.TargetFilePath)
		if err != nil {
			return err
		}
		var targets struct {
			Targets []Target `json:"targets,omitempty" yaml:"targets,omitempty"`
		}
		err = yaml.Unmarshal(bytes, &targets)
		if err != nil {
			return err
		}
		p.Targets = append(p.Targets, targets.Targets...)
	}
	if len(p.Targets) == 0 {
		return fmt.Errorf("must specify at least one target")
	}
	for _, target := range p.Targets {
		if err = validateSelector(target.Selector); err != nil {
			return err
		}
		// TODO: call validateSelector(target.NotSelector) if field added.
		if err = validateJsonFieldPath(target.FieldPath); err != nil {
			return err
		}
		if target.FilePathPosition < 0 {
			return fmt.Errorf(
				"value of FilePathPosition (%d) cannot be negative",
				target.FilePathPosition)
		}
	}
	return nil
}

// TODO: implement
func validateSelector(_ *types.Selector) error {
	return nil
}

// TODO: Enforce RFC 6902?
func validateJsonFieldPath(p string) error {
	if len(p) == 0 {
		return fmt.Errorf("fieldPath cannot be empty")
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	for _, t := range p.Targets {
		var resources []*resource.Resource
		if t.Selector == nil {
			resources = m.Resources()
		} else {
			resources, err = m.Select(*t.Selector)
			if err != nil {
				return err
			}
		}
		// TODO: consider t.NotSelector if implemented
		for _, res := range resources {
			if t.FieldPath == types.MetadataNamespacePath {
				err = res.ApplyFilter(namespace.Filter{
					Namespace: p.Value,
				})
			} else {
				err = res.ApplyFilter(valueadd.Filter{
					Value:            p.Value,
					FieldPath:        t.FieldPath,
					FilePathPosition: t.FilePathPosition,
				})
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
