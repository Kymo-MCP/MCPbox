package container

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"qm-mcp-server/pkg/k8s"
)

// KubernetesRuntime Kubernetes runtime implementation
type KubernetesRuntime struct {
	Entry *k8s.Entry
}

// NewKubernetesRuntime creates Kubernetes runtime
func NewKubernetesRuntime(kubeconfig *rest.Config, namespace string) (*KubernetesRuntime, error) {
	k8sEntry, err := k8s.NewEntry(kubeconfig, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	return &KubernetesRuntime{
		Entry: k8sEntry,
	}, nil
}

// GetContainerManager gets container manager
func (kr *KubernetesRuntime) GetContainerManager() ContainerManager {
	return &KubernetesContainerManager{Entry: kr.Entry}
}

// GetServiceManager gets service manager
func (kr *KubernetesRuntime) GetServiceManager() ServiceManager {
	return &KubernetesServiceManager{Entry: kr.Entry}
}

// GetRuntimeType gets runtime type
func (kr *KubernetesRuntime) GetRuntimeType() ContainerRuntime {
	return RuntimeKubernetes
}

// KubernetesContainerManager Kubernetes container manager implementation
type KubernetesContainerManager struct {
	Entry *k8s.Entry
}

// Create creates container (Deployment)
func (kcm *KubernetesContainerManager) Create(ctx context.Context, options ContainerCreateOptions) (string, error) {
	// Initialize basic DeploymentCreateOptions
	deploymentOptions := k8s.DeploymentCreateOptions{
		ImageName: options.ImageName,
		AppName:   options.ContainerName,
		Namespace: kcm.Entry.Namespace,
		Port:      options.Port,
		Replicas:  1, // default single replica
	}

	// Set execution command (if specified)
	if len(options.Command) > 0 {
		deploymentOptions.Command = options.Command
	}

	// Set command arguments (if specified)
	if len(options.CommandArgs) > 0 {
		deploymentOptions.CommandArgs = options.CommandArgs
	}

	// Set environment variables (if exist)
	if len(options.EnvVars) > 0 {
		deploymentOptions.EnvVars = options.EnvVars
	}

	// Set working directory (if specified)
	if options.WorkingDir != "" {
		deploymentOptions.WorkingDir = options.WorkingDir
	}

	// Set labels (if exist)
	if len(options.Labels) > 0 {
		deploymentOptions.Labels = options.Labels
	}

	// Set restart policy (if specified)
	if options.RestartPolicy != "" {
		deploymentOptions.RestartPolicy = options.RestartPolicy
	}

	// Handle volume mounts and affinity configuration
	if len(options.Mounts) > 0 {
		deploymentOptions.VolumeMounts = options.Mounts
	}

	// Set readiness probe
	if options.ReadinessProbe != nil {
		deploymentOptions.ReadinessProbe = options.ReadinessProbe
	}

	// Set image pull secrets
	if len(options.ImagePullSecrets) > 0 {
		deploymentOptions.ImagePullSecrets = options.ImagePullSecrets
	}

	// Create deployment
	deploymentName, err := kcm.Entry.Client.Deployment().Create(deploymentOptions)
	if err != nil {
		return "", err
	}

	return deploymentName, nil
}

// Delete deletes container (Deployment)
func (kcm *KubernetesContainerManager) Delete(ctx context.Context, containerName string) error {
	return kcm.Entry.Client.Deployment().Delete(containerName)
}

// Scale sets container replica count (Deployment)
func (kcm *KubernetesContainerManager) Scale(ctx context.Context, containerName string, replicas int32) error {
	return kcm.Entry.Client.Deployment().Scale(containerName, replicas)
}

// Restart restarts container (K8s environment: delete and recreate if exists, create directly if not exists)
func (kcm *KubernetesContainerManager) Restart(ctx context.Context, options ContainerCreateOptions) error {
	// Check if deployment exists
	_, err := kcm.Entry.Client.Deployment().Get(options.ContainerName)
	if err == nil {
		// deployment exists, delete first
		if err := kcm.Delete(ctx, options.ContainerName); err != nil {
			return fmt.Errorf("failed to delete existing deployment: %w", err)
		}

		// Start async process to probe and create
		go kcm.asyncProbeAndCreate(context.Background(), options)
		return nil
	} else if !isNotFoundError(err) {
		// If not NotFound error, it might be other issues (like network problems, permission issues, etc.)
		return fmt.Errorf("failed to check deployment status: %w", err)
	}

	// deployment doesn't exist, create directly
	_, createErr := kcm.Create(ctx, options)
	if createErr != nil {
		return fmt.Errorf("failed to create deployment: %w", createErr)
	}

	return nil
}

// asyncProbeAndCreate probes for deployment deletion and creates new deployment
func (kcm *KubernetesContainerManager) asyncProbeAndCreate(ctx context.Context, options ContainerCreateOptions) {
	const (
		maxProbes     = 5               // maximum probe count
		probeInterval = 5 * time.Second // probe interval
	)

	for i := 0; i < maxProbes; i++ {
		// Wait for probe interval
		time.Sleep(probeInterval)

		// Check if deployment still exists
		deployment, err := kcm.Entry.Client.Deployment().Get(options.ContainerName)
		if err != nil {
			// If get fails and is NotFound error, deletion is successful
			if isNotFoundError(err) {
				// Deployment doesn't exist, create new one
				if _, createErr := kcm.Create(ctx, options); createErr != nil {
					// Log error but don't return, as this is async
					fmt.Printf("Failed to create deployment in async process: %v\n", createErr)
				}
				return
			}
			// Other errors, continue probing
			continue
		}

		// Check if deployment is being deleted (has DeletionTimestamp)
		if deployment.DeletionTimestamp != nil {
			// Being deleted, continue probing
			continue
		}
	}

	// If we reach here, probing timed out
	fmt.Printf("Async probe timed out after %d attempts for deployment %s\n", maxProbes, options.ContainerName)
}

// waitForDeploymentDeletion waits for deployment to be completely deleted
func (kcm *KubernetesContainerManager) waitForDeploymentDeletion(ctx context.Context, deploymentName string, resourceVersion string) error {
	const (
		maxRetries    = 30              // maximum retry count
		retryInterval = 1 * time.Second // retry interval
	)

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for deployment deletion was cancelled: %w", ctx.Err())
		default:
		}

		// Check if deployment still exists
		deployment, err := kcm.Entry.Client.Deployment().Get(deploymentName)
		if err != nil {
			// If get fails and is NotFound error, deletion is successful
			if isNotFoundError(err) {
				return nil
			}
			// Other errors continue retrying
		} else {
			// If deployment still exists, check if it's a new instance (by ResourceVersion)
			if deployment.ResourceVersion != resourceVersion {
				// Different ResourceVersion means it's a newly created deployment, deletion is complete
				return nil
			}

			// Check if deployment is being deleted (has DeletionTimestamp)
			if deployment.DeletionTimestamp != nil {
				// Being deleted, continue waiting
				time.Sleep(retryInterval)
				continue
			}

			// deployment still exists and has no deletion mark, deletion might have failed
		}

		time.Sleep(retryInterval)
	}
	return fmt.Errorf("waiting for deployment deletion timed out, exceeded %d seconds", maxRetries)
}

// isNotFoundError checks if it's a NotFound error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains "not found" keyword
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "notfound")
}

// GetInfo gets container information
func (kcm *KubernetesContainerManager) GetInfo(ctx context.Context, containerName string) (*ContainerInfo, error) {
	deployment, err := kcm.Entry.Client.Deployment().Get(containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Deployment information: %w", err)
	}

	// Extract port information
	var ports []int32
	if deployment.Spec.Template.Spec.Containers != nil && len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		for _, port := range container.Ports {
			ports = append(ports, port.ContainerPort)
		}
	}

	// Determine status
	status := "Unknown"
	if deployment.Status.ReadyReplicas > 0 {
		status = "Running"
	} else if deployment.Status.Replicas > 0 {
		status = "Pending"
	} else {
		status = "Stopped"
	}

	// Get Pod IP (take the first running Pod IP)
	podIPs, _ := kcm.Entry.Client.Deployment().GetPodIPs(containerName)
	var podIP string
	if len(podIPs) > 0 {
		podIP = podIPs[0] // take the first IP
	}

	return &ContainerInfo{
		Name:      deployment.Name,
		Status:    status,
		IP:        podIP,
		Ports:     ports,
		Labels:    deployment.Labels,
		CreatedAt: deployment.CreationTimestamp.Format(time.RFC3339),
	}, nil
}

// IsReady checks if container is ready
func (kcm *KubernetesContainerManager) IsReady(ctx context.Context, containerName string) (bool, string, error) {
	ready, err := kcm.Entry.Client.Deployment().IsReady(containerName)
	if err != nil {
		return false, "check failed", err
	}

	if ready {
		return true, "ready", nil
	} else {
		return false, "not ready", nil
	}
}

// GetEvents gets container events
func (kcm *KubernetesContainerManager) GetEvents(ctx context.Context, containerName string) ([]ContainerEvent, error) {
	// Use DeploymentManager to get complete Deployment-related events
	// including Deployment, ReplicaSet and Pod events
	events, err := kcm.Entry.Client.Deployment().GetEvents(containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Deployment events: %w", err)
	}

	// Convert to generic event format
	var containerEvents []ContainerEvent
	for _, event := range events {
		containerEvents = append(containerEvents, ContainerEvent{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.FirstTimestamp.Unix(),
		})
	}

	return containerEvents, nil
}

// GetLogs gets container logs
func (kcm *KubernetesContainerManager) GetLogs(ctx context.Context, containerName string, lines int64) (string, error) {
	// Get Pod list through Deployment name
	pods, err := kcm.Entry.Client.Deployment().GetPods(containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get Pod list for Deployment: %w", err)
	}

	if len(pods) == 0 {
		return "", fmt.Errorf("no Pod found for Deployment %s", containerName)
	}

	// Get logs from the first running Pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			logs, err := kcm.Entry.Client.Pod().GetLogs(pod.Name, lines)
			if err == nil {
				return logs, nil
			}
			continue // try next Pod
		}
	}

	// If no running Pod, try to get logs from the latest Pod
	var latestPod *corev1.Pod
	for i := range pods {
		if latestPod == nil || pods[i].CreationTimestamp.After(latestPod.CreationTimestamp.Time) {
			latestPod = &pods[i]
		}
	}

	if latestPod != nil {
		logs, err := kcm.Entry.Client.Pod().GetLogs(latestPod.Name, lines)
		if err != nil {
			return "", fmt.Errorf("failed to get Pod %s logs: %w", latestPod.Name, err)
		}
		return logs, nil
	}

	return "", fmt.Errorf("no available Pod found")
}

// GetWarningEvents gets container warning events
func (kcm *KubernetesContainerManager) GetWarningEvents(ctx context.Context, containerName string) ([]ContainerEvent, error) {
	// Use DeploymentManager to get Deployment-related warning events
	// including Deployment, ReplicaSet and Pod warning events
	events, err := kcm.Entry.Client.Deployment().GetWarningEvents(containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Deployment warning events: %w", err)
	}

	// Convert to generic event format
	var containerEvents []ContainerEvent
	for _, event := range events {
		containerEvents = append(containerEvents, ContainerEvent{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.FirstTimestamp.Unix(),
		})
	}

	return containerEvents, nil
}

// KubernetesServiceManager Kubernetes service manager implementation
type KubernetesServiceManager struct {
	Entry *k8s.Entry
}

// Create creates service
func (ksm *KubernetesServiceManager) Create(ctx context.Context, serviceName string, port int32, selector map[string]string) (*ServiceInfo, error) {
	svcCfg := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None", // Headless Service
			Ports: []corev1.ServicePort{
				{
					Port: port,
				},
			},
			Selector: selector,
		},
	}

	service, err := ksm.Entry.Service.Create(svcCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service: %w", err)
	}

	// Extract port information
	var ports []int32
	for _, port := range service.Spec.Ports {
		ports = append(ports, port.Port)
	}

	return &ServiceInfo{
		Name:      service.Name,
		ClusterIP: service.Spec.ClusterIP,
		Ports:     ports,
		Labels:    service.Labels,
	}, nil
}

// Delete deletes service
func (ksm *KubernetesServiceManager) Delete(ctx context.Context, serviceName string) error {
	return ksm.Entry.Service.Delete(serviceName)
}

// Get gets service information
func (ksm *KubernetesServiceManager) Get(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	service, err := ksm.Entry.Service.Get(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Service information: %w", err)
	}

	// Extract port information
	var ports []int32
	for _, port := range service.Spec.Ports {
		ports = append(ports, port.Port)
	}

	return &ServiceInfo{
		Name:      service.Name,
		ClusterIP: service.Spec.ClusterIP,
		Ports:     ports,
		Labels:    service.Labels,
	}, nil
}

// Restart restarts service
func (ksm *KubernetesServiceManager) Restart(ctx context.Context, options ContainerCreateOptions) error {
	// Get existing service information
	_, err := ksm.Get(ctx, options.ServiceName)
	if err != nil {
		// If service does not exist, directly create new service
		if isNotFoundError(err) {
			_, createErr := ksm.Create(ctx, options.ServiceName, 80, nil)
			if createErr != nil {
				return fmt.Errorf("failed to create service: %w", createErr)
			}
			return nil
		}
		return fmt.Errorf("failed to get service information: %w", err)
	}

	// Service exists, delete first then recreate
	if err := ksm.Delete(ctx, options.ServiceName); err != nil {
		return fmt.Errorf("failed to delete existing service: %w", err)
	}

	// Wait for service to be completely deleted
	if err := ksm.waitForServiceDeletion(ctx, options.ServiceName); err != nil {
		return fmt.Errorf("failed to wait for service deletion completion: %w", err)
	}

	// Recreate service (use original port and labels, if none then use passed labels)
	_, err = ksm.Create(ctx, options.ServiceName, options.Port, options.Labels)
	if err != nil {
		return fmt.Errorf("failed to recreate service: %w", err)
	}

	return nil
}

// waitForServiceDeletion waits for service to be completely deleted
func (ksm *KubernetesServiceManager) waitForServiceDeletion(ctx context.Context, serviceName string) error {
	const (
		maxRetries    = 15              // maximum retry count
		retryInterval = 1 * time.Second // retry interval
	)

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for service deletion was cancelled: %w", ctx.Err())
		default:
		}

		// Check if service still exists
		_, err := ksm.Get(ctx, serviceName)
		if err != nil {
			// If get fails and is NotFound error, deletion is successful
			if isNotFoundError(err) {
				return nil
			}
			// Other errors continue retrying
		} else {
			// Service still exists, continue waiting
		}

		time.Sleep(retryInterval)
	}
	return fmt.Errorf("waiting for service deletion timed out, exceeded %d seconds", maxRetries)
}
