/*
Copyright The Kube Bind Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen-v0.31. DO NOT EDIT.

package v1alpha1

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"

	v1alpha1 "github.com/kube-bind/kube-bind/sdk/apis/kubebind/v1alpha1"
	kubebindv1alpha1 "github.com/kube-bind/kube-bind/sdk/kcp/applyconfiguration/kubebind/v1alpha1"
	scheme "github.com/kube-bind/kube-bind/sdk/kcp/clientset/versioned/scheme"
)

// ClusterBindingsGetter has a method to return a ClusterBindingInterface.
// A group's client should implement this interface.
type ClusterBindingsGetter interface {
	ClusterBindings(namespace string) ClusterBindingInterface
}

// ClusterBindingInterface has methods to work with ClusterBinding resources.
type ClusterBindingInterface interface {
	Create(ctx context.Context, clusterBinding *v1alpha1.ClusterBinding, opts v1.CreateOptions) (*v1alpha1.ClusterBinding, error)
	Update(ctx context.Context, clusterBinding *v1alpha1.ClusterBinding, opts v1.UpdateOptions) (*v1alpha1.ClusterBinding, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, clusterBinding *v1alpha1.ClusterBinding, opts v1.UpdateOptions) (*v1alpha1.ClusterBinding, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ClusterBinding, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ClusterBindingList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ClusterBinding, err error)
	Apply(ctx context.Context, clusterBinding *kubebindv1alpha1.ClusterBindingApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ClusterBinding, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, clusterBinding *kubebindv1alpha1.ClusterBindingApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ClusterBinding, err error)
	ClusterBindingExpansion
}

// clusterBindings implements ClusterBindingInterface
type clusterBindings struct {
	*gentype.ClientWithListAndApply[*v1alpha1.ClusterBinding, *v1alpha1.ClusterBindingList, *kubebindv1alpha1.ClusterBindingApplyConfiguration]
}

// newClusterBindings returns a ClusterBindings
func newClusterBindings(c *KubeBindV1alpha1Client, namespace string) *clusterBindings {
	return &clusterBindings{
		gentype.NewClientWithListAndApply[*v1alpha1.ClusterBinding, *v1alpha1.ClusterBindingList, *kubebindv1alpha1.ClusterBindingApplyConfiguration](
			"clusterbindings",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.ClusterBinding { return &v1alpha1.ClusterBinding{} },
			func() *v1alpha1.ClusterBindingList { return &v1alpha1.ClusterBindingList{} }),
	}
}
