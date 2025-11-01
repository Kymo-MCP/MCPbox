package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"
)

// ResourceBiz resource data processing layer
type ResourceBiz struct {
	ctx context.Context
}

// GResourceBiz global resource data processing layer instance
var GResourceBiz *ResourceBiz

func init() {
	GResourceBiz = NewResourceBiz(context.Background())
}

// NewResourceBiz create resource data processing instance
func NewResourceBiz(ctx context.Context) *ResourceBiz {
	return &ResourceBiz{
		ctx: ctx,
	}
}

// ListPVCs get PVC list by environment ID
func (biz *ResourceBiz) ListPVCs(environmentID uint) ([]k8s.PVCInfo, error) {
	// Get environment configuration
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8s client: %s", err.Error())
	}

	// Call Volume manager to get PVC list, specify environment namespace
	return k8sEntry.Volume.ListPVCs(k8sEntry.Namespace)
}

// ListNodes get node list by environment ID
func (biz *ResourceBiz) ListNodes(environmentID uint) ([]k8s.NodeInfo, error) {
	// Get environment configuration
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8s client: %s", err.Error())
	}

	// Call Node manager to get node list
	return k8sEntry.Node.ListNodes()
}

// ListStorageClasses get storage class list by environment ID
func (biz *ResourceBiz) ListStorageClasses(environmentID uint) ([]k8s.StorageClassInfo, error) {
	// Get environment configuration
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8s client: %s", err.Error())
	}

	// Call Volume manager to get storage class list
	return k8sEntry.Volume.ListStorageClasses()
}

// CreateHostPathPVC create host path based PVC by environment ID
func (biz *ResourceBiz) CreateHostPathPVC(environmentID uint, name, hostPath, nodeName, accessMode, storageClass string, storageSize int32) (*k8s.PVCInfo, error) {
	// Get environment configuration
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8s client: %s", err.Error())
	}

	// Call Volume manager to create host path based PVC and PV
	// Use empty label map as labels are no longer needed
	return k8sEntry.Volume.CreateHostPathPVCWithPV(name, hostPath, nodeName, accessMode, storageSize, storageClass, nil)
}

// CreatePVC create regular PVC by environment ID
func (biz *ResourceBiz) CreatePVC(environmentID uint, name, nodeName, accessMode, storageClass string, storageSize int32, labels map[string]string) (*k8s.PVCInfo, error) {
	// Get environment configuration
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8s client: %s", err.Error())
	}

	// Call Volume manager to create regular PVC
	return k8sEntry.Volume.CreatePVCWithParams(name, nodeName, accessMode, storageClass, storageSize, labels)
}

// getK8sEntryByEnvironmentID get K8s Entry by environment ID
func (biz *ResourceBiz) getK8sEntryByEnvironmentID(environmentID uint) (*k8s.Entry, error) {
	// Get environment information
	environment, err := GEnvironmentBiz.GetEnvironment(biz.ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment information: %s", err.Error())
	}

	// Validate environment type
	if environment.Environment != model.McpEnvironmentKubernetes {
		return nil, fmt.Errorf("environment type is not Kubernetes, cannot query K8s resources")
	}

	// Create container runtime configuration
	config := container.Config{
		Runtime:    container.RuntimeKubernetes,
		Namespace:  environment.Namespace,
		Kubeconfig: common.SetKubeConfig([]byte(environment.Config)),
		Network:    "bridge", // Default network configuration
	}

	// Create container runtime entry
	entry, err := container.NewEntry(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container runtime entry: %s", err.Error())
	}

	// Check if it's Kubernetes runtime
	if !entry.IsKubernetes() {
		return nil, fmt.Errorf("runtime type error, expected Kubernetes runtime")
	}

	// Get K8s entry
	k8sRuntime := entry.GetK8sRuntime()
	if k8sRuntime == nil {
		return nil, fmt.Errorf("failed to get K8s entry")
	}
	if k8sEntry := k8sRuntime.Entry; k8sEntry != nil {
		return k8sEntry, nil
	}

	return nil, fmt.Errorf("K8s runtime type assertion failed")
}
