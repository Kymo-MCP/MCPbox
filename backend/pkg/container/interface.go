package container

import (
	"context"

	"github.com/kymo-mcp/mcpcan/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
)

// ContainerRuntime container runtime type
type ContainerRuntime string

const (
	// RuntimeKubernetes Kubernetes runtime
	RuntimeKubernetes ContainerRuntime = "kubernetes"
	// RuntimeDocker Docker runtime
	RuntimeDocker ContainerRuntime = "docker"
)

// MountType mount type enumeration
type MountType string

const (
	// MountTypeHostPath HostPath type mount
	MountTypeHostPath MountType = "hostPath"
	// MountTypePVC PVC type mount
	MountTypePVC MountType = "pvc"
	// MountTypeConfigMap ConfigMap type mount
	MountTypeConfigMap MountType = "configMap"
)

// VolumeMount volume mount configuration (deprecated, use UnifiedMount instead)
// Deprecated: use UnifiedMount instead
type VolumeMount struct {
	Type       string `json:"type"`       // "hostPath" or "pvc"
	SourcePath string `json:"sourcePath"` // host path or PVC name
	TargetPath string `json:"targetPath"` // mount path in container
	NodeID     string `json:"nodeID"`     // node ID (only required for host mount)
	ReadOnly   bool   `json:"readOnly"`   // read-only flag
}

// ConfigMapMount ConfigMap mount configuration (deprecated, use UnifiedMount instead)
// Deprecated: use UnifiedMount instead
type ConfigMapMount struct {
	ConfigMapName string `json:"configMapName"` // ConfigMap name
	MountPath     string `json:"mountPath"`     // mount path
}

// ContainerCreateOptions container creation options
type ContainerCreateOptions struct {
	ImageName        string             `json:"imageName"`        // image name
	ContainerName    string             `json:"containerName"`    // container name
	ServiceName      string             `json:"serviceName"`      // service name
	Port             int32              `json:"port"`             // port
	Command          []string           `json:"command"`          // execution command (overrides image ENTRYPOINT, Docker: --entrypoint, K8s: command)
	CommandArgs      []string           `json:"commandArgs"`      // command arguments (overrides image CMD, Docker: args after image, K8s: args)
	EnvVars          map[string]string  `json:"envVars"`          // environment variables
	Mounts           []k8s.UnifiedMount `json:"mounts"`           // volume mounts
	ReadinessProbe   *corev1.Probe      `json:"readinessProbe"`   // readiness probe
	Labels           map[string]string  `json:"labels"`           // labels
	RestartPolicy    string             `json:"restartPolicy"`    // restart policy (Docker: no/always/unless-stopped/on-failure)
	WorkingDir       string             `json:"workingDir"`       // working directory
	ImagePullSecrets []string           `json:"imagePullSecrets"` // image pull secret names list (only applicable to Kubernetes)

}

// ContainerInfo container information
type ContainerInfo struct {
	Name      string            // container name
	Status    string            // container status
	IP        string            // container IP
	Ports     []int32           // port list
	Labels    map[string]string // labels
	CreatedAt string            // creation time
}

// ServiceInfo service information
type ServiceInfo struct {
	Name      string            // service name
	ClusterIP string            // cluster IP
	Ports     []int32           // port list
	Labels    map[string]string // labels
}

// ContainerEvent container event
type ContainerEvent struct {
	Type      string // event type
	Reason    string // event reason
	Message   string // event message
	Timestamp int64  // timestamp
}

// ContainerManager container manager interface
type ContainerManager interface {
	// Create creates a container
	Create(ctx context.Context, options ContainerCreateOptions) (string, error)
	// Delete deletes a container
	Delete(ctx context.Context, containerName string) error
	// Scale sets container replica count (only applicable to k8s)
	Scale(ctx context.Context, containerName string, replicas int32) error
	// Restart restarts a container
	Restart(ctx context.Context, options ContainerCreateOptions) error
	// GetInfo gets container information
	GetInfo(ctx context.Context, containerName string) (*ContainerInfo, error)
	// IsReady checks if container is ready
	IsReady(ctx context.Context, containerName string) (bool, string, error)
	// GetEvents gets container events
	GetEvents(ctx context.Context, containerName string) ([]ContainerEvent, error)
	// GetWarningEvents gets container warning events
	GetWarningEvents(ctx context.Context, containerName string) ([]ContainerEvent, error)
	// GetLogs gets container logs
	GetLogs(ctx context.Context, containerName string, lines int64) (string, error)
}

// ServiceManager service manager interface
type ServiceManager interface {
	// Create creates a service
	Create(ctx context.Context, serviceName string, port int32, selector map[string]string) (*ServiceInfo, error)
	// Delete deletes a service
	Delete(ctx context.Context, serviceName string) error
	// Get gets service information
	Get(ctx context.Context, serviceName string) (*ServiceInfo, error)
	// Restart restarts a service
	Restart(ctx context.Context, options ContainerCreateOptions) error
}

// ContainerRuntime container runtime interface
type Runtime interface {
	// GetContainerManager gets container manager
	GetContainerManager() ContainerManager
	// GetServiceManager gets service manager
	GetServiceManager() ServiceManager
	// GetRuntimeType gets runtime type
	GetRuntimeType() ContainerRuntime
}
