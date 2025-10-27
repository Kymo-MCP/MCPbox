package container

import (
	"fmt"

	"k8s.io/client-go/rest"
)

// Entry unified entry point for container runtime
type Entry struct {
	runtime Runtime
	config  Config
}

// Config container runtime configuration
type Config struct {
	Runtime    ContainerRuntime `yaml:"runtime" json:"runtime"`       // runtime type: kubernetes or docker
	Namespace  string           `yaml:"namespace" json:"namespace"`   // Kubernetes namespace
	Kubeconfig *rest.Config     `yaml:"kubeconfig" json:"kubeconfig"` // Kubernetes configuration file path
	Network    string           `yaml:"network" json:"network"`       // Docker network name
}

// NewEntry creates container runtime entry
func NewEntry(config Config) (*Entry, error) {
	var runtime Runtime
	var err error

	switch config.Runtime {
	case RuntimeKubernetes:
		runtime, err = NewKubernetesRuntime(config.Kubeconfig, config.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Kubernetes runtime: %w", err)
		}
	case RuntimeDocker:
		runtime = NewDockerRuntime(config.Network)
	default:
		return nil, fmt.Errorf("unsupported container runtime: %s", config.Runtime)
	}

	return &Entry{
		runtime: runtime,
		config:  config,
	}, nil
}

// GetRuntime gets container runtime
func (e *Entry) GetRuntime() Runtime {
	return e.runtime
}

// GetConfig gets configuration
func (e *Entry) GetConfig() Config {
	return e.config
}

// GetContainerManager gets container manager
func (e *Entry) GetContainerManager() ContainerManager {
	return e.runtime.GetContainerManager()
}

// GetServiceManager gets service manager
func (e *Entry) GetServiceManager() ServiceManager {
	return e.runtime.GetServiceManager()
}

// GetRuntimeType gets runtime type
func (e *Entry) GetRuntimeType() ContainerRuntime {
	return e.runtime.GetRuntimeType()
}

// SwitchRuntime switches container runtime (dynamic switching)
func (e *Entry) SwitchRuntime(config Config) error {
	var runtime Runtime
	var err error

	switch config.Runtime {
	case RuntimeKubernetes:
		runtime, err = NewKubernetesRuntime(config.Kubeconfig, config.Namespace)
		if err != nil {
			return fmt.Errorf("failed to switch to Kubernetes runtime: %w", err)
		}
	case RuntimeDocker:
		runtime = NewDockerRuntime(config.Network)
	default:
		return fmt.Errorf("unsupported container runtime: %s", config.Runtime)
	}

	e.runtime = runtime
	e.config = config
	return nil
}

// IsKubernetes checks if current runtime is Kubernetes
func (e *Entry) IsKubernetes() bool {
	return e.config.Runtime == RuntimeKubernetes
}

// IsDocker checks if current runtime is Docker
func (e *Entry) IsDocker() bool {
	return e.config.Runtime == RuntimeDocker
}

// GetK8sRuntime gets K8s runtime Entry (if current runtime is Kubernetes)
func (e *Entry) GetK8sRuntime() *KubernetesRuntime {
	if kr, ok := e.runtime.(*KubernetesRuntime); ok {
		return kr
	}
	return nil
}

// ListNamespaces lists all namespaces
func (e *Entry) ListNamespaces() ([]string, error) {
	if !e.IsKubernetes() {
		return nil, fmt.Errorf("runtime is not Kubernetes")
	}

	k8sRuntime := e.GetK8sRuntime()
	if k8sRuntime == nil {
		return nil, fmt.Errorf("failed to get Kubernetes entry")
	}

	namespaces, err := k8sRuntime.Entry.Client.ListNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	return namespaces, nil
}
