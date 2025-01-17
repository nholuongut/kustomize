// Copyright 2019 Nho Luong DevOps.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/yaml"
)

// Sort the resources using a customizable ordering based of Kind.
// Defaults to the ordering of the GVK struct, which puts cluster-wide basic
// resources with no dependencies (like Namespace, StorageClass, etc.) first,
// and resources with a high number of dependencies
// (like ValidatingWebhookConfiguration) last.
type plugin struct {
	SortOptions *types.SortOptions `json:"sortOptions,omitempty" yaml:"sortOptions,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) error {
	return errors.WrapPrefixf(yaml.Unmarshal(c, p), "Failed to unmarshal SortOrderTransformer config")
}

func (p *plugin) applyDefaults() {
	// Default to FIFO sort, aka no-op.
	if p.SortOptions == nil {
		p.SortOptions = &types.SortOptions{
			Order: types.FIFOSortOrder,
		}
	}

	// If legacy sort is selected and no options are given, default to
	// hardcoded order.
	if p.SortOptions.Order == types.LegacySortOrder && p.SortOptions.LegacySortOptions == nil {
		p.SortOptions.LegacySortOptions = &types.LegacySortOptions{
			OrderFirst: defaultOrderFirst,
			OrderLast:  defaultOrderLast,
		}
	}
}

func (p *plugin) validate() error {
	// Check valid values for SortOrder
	if p.SortOptions.Order != types.FIFOSortOrder && p.SortOptions.Order != types.LegacySortOrder {
		return errors.Errorf("the field 'sortOptions.order' must be one of [%s, %s]",
			types.FIFOSortOrder, types.LegacySortOrder)
	}

	// Validate that the only options set are the ones corresponding to the
	// selected sort order.
	if p.SortOptions.Order == types.FIFOSortOrder &&
		p.SortOptions.LegacySortOptions != nil {
		return errors.Errorf("the field 'sortOptions.legacySortOptions' is"+
			" set but the selected sort order is '%v', not 'legacy'",
			p.SortOptions.Order)
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	p.applyDefaults()
	err = p.validate()
	if err != nil {
		return err
	}

	// Sort
	if p.SortOptions.Order == types.LegacySortOrder {
		s := newLegacyIDSorter(m.Resources(), p.SortOptions.LegacySortOptions)
		sort.Sort(s)

		// Clear the map and re-add the resources in the sorted order.
		m.Clear()
		for _, r := range s.resources {
			err := m.Append(r)
			if err != nil {
				return errors.WrapPrefixf(err, "SortOrderTransformer: Failed to append to resources")
			}
		}
	}
	return nil
}

// Code for legacy sorting.
// Legacy sorting is a "fixed" order sorting maintained for backwards
// compatibility.

// legacyIDSorter sorts resources based on two priority lists:
// - orderFirst: Resources that should be placed in the start, in the given order.
// - orderLast: Resources that should be placed in the end, in the given order.
type legacyIDSorter struct {
	// resids only stores the metadata of the object. This is an optimization as
	// it's expensive to compute these again and again during ordering.
	resids []resid.ResId
	// Initially, we sorted the metadata (ResId) of each object and then called GetByCurrentId on each to construct the final list.
	// The problem is that GetByCurrentId is inefficient and does a linear scan in a list every time we do that.
	// So instead, we sort resources alongside the ResIds.
	resources []*resource.Resource

	typeOrders map[string]int
}

func newLegacyIDSorter(
	resources []*resource.Resource,
	options *types.LegacySortOptions) *legacyIDSorter {
	// Precalculate a resource ranking based on the priority lists.
	var typeOrders = func() map[string]int {
		m := map[string]int{}
		for i, n := range options.OrderFirst {
			m[n] = -len(options.OrderFirst) + i
		}
		for i, n := range options.OrderLast {
			m[n] = 1 + i
		}
		return m
	}()

	ret := &legacyIDSorter{typeOrders: typeOrders}
	for _, res := range resources {
		ret.resids = append(ret.resids, res.CurId())
		ret.resources = append(ret.resources, res)
	}
	return ret
}

var _ sort.Interface = legacyIDSorter{}

func (a legacyIDSorter) Len() int { return len(a.resids) }
func (a legacyIDSorter) Swap(i, j int) {
	a.resids[i], a.resids[j] = a.resids[j], a.resids[i]
	a.resources[i], a.resources[j] = a.resources[j], a.resources[i]
}
func (a legacyIDSorter) Less(i, j int) bool {
	if !a.resids[i].Gvk.Equals(a.resids[j].Gvk) {
		return gvkLessThan(a.resids[i].Gvk, a.resids[j].Gvk, a.typeOrders)
	}
	return legacyResIDSortString(a.resids[i]) < legacyResIDSortString(a.resids[j])
}

func gvkLessThan(gvk1, gvk2 resid.Gvk, typeOrders map[string]int) bool {
	index1 := typeOrders[gvk1.Kind]
	index2 := typeOrders[gvk2.Kind]
	if index1 != index2 {
		return index1 < index2
	}
	if (gvk1.Kind == types.NamespaceKind && gvk2.Kind == types.NamespaceKind) && (gvk1.Group == "" || gvk2.Group == "") {
		return legacyGVKSortString(gvk1) > legacyGVKSortString(gvk2)
	}
	return legacyGVKSortString(gvk1) < legacyGVKSortString(gvk2)
}

// legacyGVKSortString returns a string representation of given GVK used for
// stable sorting.
func legacyGVKSortString(x resid.Gvk) string {
	legacyNoGroup := "~G"
	legacyNoVersion := "~V"
	legacyNoKind := "~K"
	legacyFieldSeparator := "_"

	g := x.Group
	if g == "" {
		g = legacyNoGroup
	}
	v := x.Version
	if v == "" {
		v = legacyNoVersion
	}
	k := x.Kind
	if k == "" {
		k = legacyNoKind
	}
	return strings.Join([]string{g, v, k}, legacyFieldSeparator)
}

// legacyResIDSortString returns a string representation of given ResID used for
// stable sorting.
func legacyResIDSortString(id resid.ResId) string {
	legacyNoNamespace := "~X"
	legacyNoName := "~N"
	legacySeparator := "|"

	ns := id.Namespace
	if ns == "" {
		ns = legacyNoNamespace
	}
	nm := id.Name
	if nm == "" {
		nm = legacyNoName
	}
	return strings.Join(
		[]string{id.Gvk.String(), ns, nm}, legacySeparator)
}

// DO NOT CHANGE!
// Final legacy ordering provided as a default by kustomize.
// Originally an attempt to apply resources in the correct order, an effort
// which later proved impossible as not all types are known beforehand.
// See: https://github.com/nholuongut/kustomize/issues/3913
var defaultOrderFirst = []string{ //nolint:gochecknoglobals
	"Namespace",
	"ResourceQuota",
	"StorageClass",
	"CustomResourceDefinition",
	"ServiceAccount",
	"PodSecurityPolicy",
	"Role",
	"ClusterRole",
	"RoleBinding",
	"ClusterRoleBinding",
	"ConfigMap",
	"Secret",
	"Endpoints",
	"Service",
	"LimitRange",
	"PriorityClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"Deployment",
	"StatefulSet",
	"CronJob",
	"PodDisruptionBudget",
}
var defaultOrderLast = []string{ //nolint:gochecknoglobals
	"MutatingWebhookConfiguration",
	"ValidatingWebhookConfiguration",
}
