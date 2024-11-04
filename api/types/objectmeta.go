// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

package types

// ObjectMeta partially copies apimachinery/pkg/apis/meta/v1.ObjectMeta
// No need for a direct dependence; the fields are stable.
type ObjectMeta struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}
