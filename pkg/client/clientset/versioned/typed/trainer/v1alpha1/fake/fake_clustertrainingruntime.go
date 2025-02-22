// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "github.com/kubeflow/trainer/pkg/apis/trainer/v1alpha1"
	trainerv1alpha1 "github.com/kubeflow/trainer/pkg/client/applyconfiguration/trainer/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterTrainingRuntimes implements ClusterTrainingRuntimeInterface
type FakeClusterTrainingRuntimes struct {
	Fake *FakeTrainerV1alpha1
}

var clustertrainingruntimesResource = v1alpha1.SchemeGroupVersion.WithResource("clustertrainingruntimes")

var clustertrainingruntimesKind = v1alpha1.SchemeGroupVersion.WithKind("ClusterTrainingRuntime")

// Get takes name of the clusterTrainingRuntime, and returns the corresponding clusterTrainingRuntime object, and an error if there is any.
func (c *FakeClusterTrainingRuntimes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ClusterTrainingRuntime, err error) {
	emptyResult := &v1alpha1.ClusterTrainingRuntime{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(clustertrainingruntimesResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ClusterTrainingRuntime), err
}

// List takes label and field selectors, and returns the list of ClusterTrainingRuntimes that match those selectors.
func (c *FakeClusterTrainingRuntimes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterTrainingRuntimeList, err error) {
	emptyResult := &v1alpha1.ClusterTrainingRuntimeList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(clustertrainingruntimesResource, clustertrainingruntimesKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ClusterTrainingRuntimeList{ListMeta: obj.(*v1alpha1.ClusterTrainingRuntimeList).ListMeta}
	for _, item := range obj.(*v1alpha1.ClusterTrainingRuntimeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterTrainingRuntimes.
func (c *FakeClusterTrainingRuntimes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(clustertrainingruntimesResource, opts))
}

// Create takes the representation of a clusterTrainingRuntime and creates it.  Returns the server's representation of the clusterTrainingRuntime, and an error, if there is any.
func (c *FakeClusterTrainingRuntimes) Create(ctx context.Context, clusterTrainingRuntime *v1alpha1.ClusterTrainingRuntime, opts v1.CreateOptions) (result *v1alpha1.ClusterTrainingRuntime, err error) {
	emptyResult := &v1alpha1.ClusterTrainingRuntime{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(clustertrainingruntimesResource, clusterTrainingRuntime, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ClusterTrainingRuntime), err
}

// Update takes the representation of a clusterTrainingRuntime and updates it. Returns the server's representation of the clusterTrainingRuntime, and an error, if there is any.
func (c *FakeClusterTrainingRuntimes) Update(ctx context.Context, clusterTrainingRuntime *v1alpha1.ClusterTrainingRuntime, opts v1.UpdateOptions) (result *v1alpha1.ClusterTrainingRuntime, err error) {
	emptyResult := &v1alpha1.ClusterTrainingRuntime{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(clustertrainingruntimesResource, clusterTrainingRuntime, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ClusterTrainingRuntime), err
}

// Delete takes name of the clusterTrainingRuntime and deletes it. Returns an error if one occurs.
func (c *FakeClusterTrainingRuntimes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(clustertrainingruntimesResource, name, opts), &v1alpha1.ClusterTrainingRuntime{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterTrainingRuntimes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(clustertrainingruntimesResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ClusterTrainingRuntimeList{})
	return err
}

// Patch applies the patch and returns the patched clusterTrainingRuntime.
func (c *FakeClusterTrainingRuntimes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ClusterTrainingRuntime, err error) {
	emptyResult := &v1alpha1.ClusterTrainingRuntime{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(clustertrainingruntimesResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ClusterTrainingRuntime), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied clusterTrainingRuntime.
func (c *FakeClusterTrainingRuntimes) Apply(ctx context.Context, clusterTrainingRuntime *trainerv1alpha1.ClusterTrainingRuntimeApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ClusterTrainingRuntime, err error) {
	if clusterTrainingRuntime == nil {
		return nil, fmt.Errorf("clusterTrainingRuntime provided to Apply must not be nil")
	}
	data, err := json.Marshal(clusterTrainingRuntime)
	if err != nil {
		return nil, err
	}
	name := clusterTrainingRuntime.Name
	if name == nil {
		return nil, fmt.Errorf("clusterTrainingRuntime.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.ClusterTrainingRuntime{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(clustertrainingruntimesResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ClusterTrainingRuntime), err
}
