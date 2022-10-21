/*
Copyright 2022 The Kube Bind Authors.

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

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	kuberesources "github.com/kube-bind/kube-bind/contrib/example-backend/kubernetes/resources"
	bindclient "github.com/kube-bind/kube-bind/pkg/client/clientset/versioned"
	bindinformers "github.com/kube-bind/kube-bind/pkg/client/informers/externalversions/kubebind/v1alpha1"
	bindlisters "github.com/kube-bind/kube-bind/pkg/client/listers/kubebind/v1alpha1"
	"github.com/kube-bind/kube-bind/pkg/indexers"
)

type Manager struct {
	namespacePrefix    string
	providerPrettyName string

	clusterConfig *rest.Config

	kubeClient kubeclient.Interface
	bindClient bindclient.Interface

	namespaceLister  corev1listers.NamespaceLister
	namespaceIndexer cache.Indexer

	exportLister  bindlisters.APIServiceExportLister
	exportIndexer cache.Indexer
}

func NewKubernetesManager(
	namespacePrefix, providerPrettyName string,
	config *rest.Config,
	namespaceInformer corev1informers.NamespaceInformer,
	exportInformer bindinformers.APIServiceExportInformer,
) (*Manager, error) {
	config = rest.CopyConfig(config)
	config = rest.AddUserAgent(config, "kube-bind-example-backend-kubernetes-manager")

	kubeClient, err := kubeclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	bindClient, err := bindclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		namespacePrefix:    namespacePrefix,
		providerPrettyName: providerPrettyName,

		clusterConfig: config,

		kubeClient: kubeClient,
		bindClient: bindClient,

		namespaceLister:  namespaceInformer.Lister(),
		namespaceIndexer: namespaceInformer.Informer().GetIndexer(),

		exportLister:  exportInformer.Lister(),
		exportIndexer: exportInformer.Informer().GetIndexer(),
	}

	indexers.AddIfNotPresentOrDie(m.namespaceIndexer, cache.Indexers{
		NamespacesByIdentity: IndexNamespacesByIdentity,
	})

	indexers.AddIfNotPresentOrDie(m.namespaceIndexer, cache.Indexers{
		indexers.ServiceExportByServiceExportResource: indexers.IndexServiceExportByServiceExportResource,
	})

	return m, nil
}

func (m *Manager) HandleResources(ctx context.Context, identity, resource, group string) ([]byte, error) {
	logger := klog.FromContext(ctx).WithValues("identity", identity, "resource", resource, "group", group)
	ctx = klog.NewContext(ctx, logger)

	// try to find an existing namespace by annotation, or create a new one.
	nss, err := m.namespaceIndexer.ByIndex(NamespacesByIdentity, identity)
	if err != nil {
		return nil, err
	}
	if len(nss) > 1 {
		logger.Error(fmt.Errorf("found multiple namespaces for identity %q", identity), "found multiple namespaces for identity")
		return nil, fmt.Errorf("found multiple namespaces for identity %q", identity)
	}
	var ns string
	if len(nss) == 1 {
		ns = nss[0].(*corev1.Namespace).Name
	} else {
		nsObj, err := kuberesources.CreateNamespace(ctx, m.kubeClient, m.namespacePrefix, identity)
		logger.Info("Created namespace", "namespace", nsObj.Name)
		if err != nil {
			return nil, err
		}
		ns = nsObj.Name
	}
	logger = logger.WithValues("namespace", ns)
	ctx = klog.NewContext(ctx, logger)

	// first look for ClusterBinding to get old secret name
	kubeconfigSecretName := kuberesources.KubeconfigSecretName
	cb, err := m.bindClient.KubeBindV1alpha1().ClusterBindings(ns).Get(ctx, kuberesources.ClusterBindingName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		if err := kuberesources.CreateClusterBinding(ctx, m.bindClient, ns, "kubeconfig", m.providerPrettyName); err != nil {
			return nil, err
		}
	} else {
		logger.V(3).Info("Found existing ClusterBinding")
		kubeconfigSecretName = cb.Spec.KubeconfigSecretRef.Name // reuse old name
	}

	sa, err := kuberesources.CreateServiceAccount(ctx, m.kubeClient, ns, kuberesources.ServiceAccountName)
	if err != nil {
		return nil, err
	}

	saSecret, err := kuberesources.CreateSASecret(ctx, m.kubeClient, ns, sa.Name)
	if err != nil {
		return nil, err
	}

	kfgSecret, err := kuberesources.GenerateKubeconfig(ctx, m.kubeClient, m.clusterConfig, saSecret.Name, ns, kubeconfigSecretName)
	if err != nil {
		return nil, err
	}

	return kfgSecret.Data["kubeconfig"], nil
}
