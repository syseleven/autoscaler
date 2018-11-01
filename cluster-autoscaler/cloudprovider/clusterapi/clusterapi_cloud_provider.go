/*
Copyright 2018 The Kubernetes Authors.

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

package clusterapi

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const (
	// ProviderName is the cloud provider name for ClusterApi
	ProviderName = "clusterapi"
)

// ClusterapiCloudProvider should have comment or be unexported
type ClusterapiCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	machineManager  MachineManager
}

// BuildClusterapiCloudProvider creates new ClusterapiCloudProvider
func BuildClusterapiCloudProvider(machineManager MachineManager, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	if err := machineManager.Refresh(); err != nil {
		return nil, err
	}

	clusterapi := &ClusterapiCloudProvider{
		resourceLimiter: resourceLimiter,
		machineManager:  machineManager,
	}

	return clusterapi, nil
}

// Name returns name of the cloud provider.
func (clusterapi *ClusterapiCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (clusterapi *ClusterapiCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	mds := clusterapi.machineManager.AllDeployments()
	ngs := make([]cloudprovider.NodeGroup, len(mds))
	for i, md := range mds {
		ngs[i] = NewClusterapiNodeGroup(clusterapi.machineManager, md)
	}

	return ngs
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred.
func (clusterapi *ClusterapiCloudProvider) NodeGroupForNode(node *v1.Node) (cloudprovider.NodeGroup, error) {
	if md := clusterapi.machineManager.DeploymentForNode(node); md != nil {
		return NewClusterapiNodeGroup(clusterapi.machineManager, md), nil
	}
	// node is not part of a nodegroup, this is perfectly fine just return nil
	return nil, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (clusterapi *ClusterapiCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (clusterapi *ClusterapiCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (clusterapi *ClusterapiCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []v1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (clusterapi *ClusterapiCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return clusterapi.resourceLimiter, nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (clusterapi *ClusterapiCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (clusterapi *ClusterapiCloudProvider) Refresh() error {
	return clusterapi.machineManager.Refresh()
}

// BuildClusterapi builds Clusterapi cloud provider, manager etc.
func BuildClusterapi(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, kubeConfig *rest.Config) cloudprovider.CloudProvider {
	machineManager, err := NewMachineManager(kubeConfig)
	if err != nil {
		klog.Fatalf("Failed to create Clusterapi machine manager: %v", err)
	}
	provider, err := BuildClusterapiCloudProvider(machineManager, rl)
	if err != nil {
		klog.Fatalf("Failed to create Clusterapi cloud provider: %v", err)
	}
	return provider
}