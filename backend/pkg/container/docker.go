package container

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/k8s"
)

// DockerRuntime Docker runtime implementation
type DockerRuntime struct {
	networkName string // Docker network name
}

// NewDockerRuntime creates Docker runtime
func NewDockerRuntime(networkName string) *DockerRuntime {
	if networkName == "" {
		networkName = "bridge" // default network
	}
	return &DockerRuntime{
		networkName: networkName,
	}
}

// GetContainerManager gets container manager
func (dr *DockerRuntime) GetContainerManager() ContainerManager {
	return &DockerContainerManager{networkName: dr.networkName}
}

// GetServiceManager gets service manager
func (dr *DockerRuntime) GetServiceManager() ServiceManager {
	return &DockerServiceManager{
		networkName: dr.networkName,
	}
}

// GetRuntimeType gets runtime type
func (dr *DockerRuntime) GetRuntimeType() ContainerRuntime {
	return RuntimeDocker
}

// DockerContainerManager Docker container manager implementation
type DockerContainerManager struct {
	networkName string
}

// DockerContainerInfo Docker container information structure
type DockerContainerInfo struct {
	ID      string            `json:"ID"`
	Names   []string          `json:"names"`
	Image   string            `json:"image"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Ports   []DockerPort      `json:"ports"`
	Labels  map[string]string `json:"labels"`
	Created int64             `json:"created"`
}

// DockerPort Docker port information
type DockerPort struct {
	PrivatePort int32  `json:"privatePort"`
	PublicPort  int32  `json:"publicPort"`
	Type        string `json:"type"`
}

// Create creates container
func (dcm *DockerContainerManager) Create(ctx context.Context, options ContainerCreateOptions) (string, error) {
	// Build docker run command
	args := []string{"run", "-d"}

	// Set container name
	if options.ContainerName != "" {
		args = append(args, "--name", options.ContainerName)
	}

	// Set network
	if dcm.networkName != "" {
		args = append(args, "--network", dcm.networkName)
	}

	// Set restart policy
	if options.RestartPolicy != "" {
		// Validate restart policy
		validPolicies := []string{"no", "on-failure", "always", "unless-stopped"}
		isValid := false
		for _, policy := range validPolicies {
			if options.RestartPolicy == policy {
				isValid = true
				break
			}
		}
		if !isValid {
			return "", fmt.Errorf("invalid restart policy: %s", options.RestartPolicy)
		}
		args = append(args, "--restart", options.RestartPolicy)
	}

	// Set working directory
	if options.WorkingDir != "" {
		// Ensure working directory is absolute path
		if !strings.HasPrefix(options.WorkingDir, "/") {
			options.WorkingDir = "/" + options.WorkingDir
		}
		args = append(args, "-w", options.WorkingDir)
	}

	// Set port mapping
	if options.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", options.Port, options.Port))
	}

	// Set environment variables
	for key, value := range options.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Set volume mounts
	for _, mount := range options.Mounts {
		if mount.Type == k8s.MountTypeHostPath {
			// For Docker, treat HostPath as bind mount
			args = append(args, "--mount", fmt.Sprintf("type=bind,source=%s,target=%s", mount.HostPath, mount.MountPath))
		} else if mount.Type == k8s.MountTypePVC {
			// For Docker, treat PVC as volume mount
			args = append(args, "-v", fmt.Sprintf("%s:%s", mount.PVCName, mount.MountPath))
		}
	}

	// Set labels
	for key, value := range options.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	// Set health check
	if options.ReadinessProbe != nil {
		if options.ReadinessProbe.HTTPGet != nil {
			healthCmd := fmt.Sprintf("curl -f http://localhost:%d%s || exit 1",
				options.ReadinessProbe.HTTPGet.Port,
				options.ReadinessProbe.HTTPGet.Path)
			args = append(args, "--health-cmd", healthCmd)
			args = append(args, "--health-interval", "30s")
			args = append(args, "--health-timeout", "3s")
			args = append(args, "--health-retries", "3")
		}
	}

	// Set entry point program (overrides image ENTRYPOINT, must be before image name)
	if len(options.Command) > 0 {
		args = append(args, "--entrypoint", options.Command[0])
	}

	// Add image name
	args = append(args, options.ImageName)

	// Add command arguments (overrides image CMD)
	if len(options.CommandArgs) > 0 {
		args = append(args, options.CommandArgs...)
	}

	// Execute docker run command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create Docker container: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Delete deletes container
func (dcm *DockerContainerManager) Delete(ctx context.Context, containerName string) error {
	// Stop container
	stopCmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	_ = stopCmd.Run() // ignore stop error, container might already be stopped

	// Delete container
	deleteCmd := exec.CommandContext(ctx, "docker", "rm", containerName)
	if err := deleteCmd.Run(); err != nil {
		return fmt.Errorf("failed to delete Docker container: %w", err)
	}
	return nil
}

// Scale sets container replica count (in Docker environment, 0 means delete, greater than 0 is not supported)
func (dcm *DockerContainerManager) Scale(ctx context.Context, containerName string, replicas int32) error {
	if replicas == 0 {
		return dcm.Delete(ctx, containerName)
	}
	return fmt.Errorf("Docker environment does not support setting replica count to %d, only supports 0 (delete container)", replicas)
}

// Restart restarts container (Docker environment: delete and restart if exists, start directly if not exists)
func (dcm *DockerContainerManager) Restart(ctx context.Context, options ContainerCreateOptions) error {
	// Check if container exists
	_, err := dcm.GetInfo(ctx, options.ContainerName)
	if err == nil {
		// Container exists, delete first
		if err := dcm.Delete(ctx, options.ContainerName); err != nil {
			return fmt.Errorf("failed to delete existing container: %w", err)
		}
	} else if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "No such container") {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	// Create new container
	_, err = dcm.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	return nil
}

// GetInfo gets container information
func (dcm *DockerContainerManager) GetInfo(ctx context.Context, containerName string) (*ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{json .}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker container information: %w", err)
	}

	var dockerInfo DockerContainerInfo
	if err := json.Unmarshal(output, &dockerInfo); err != nil {
		return nil, fmt.Errorf("failed to parse Docker container information: %w", err)
	}

	if dockerInfo.ID == "" {
		return nil, fmt.Errorf("container does not exist: %s", containerName)
	}

	// Extract port information
	var ports []int32
	for _, port := range dockerInfo.Ports {
		ports = append(ports, port.PrivatePort)
	}

	// Get container IP
	ip, _ := dcm.getContainerIP(ctx, containerName)

	return &ContainerInfo{
		Name:      strings.TrimPrefix(dockerInfo.Names[0], "/"),
		Status:    dockerInfo.State,
		IP:        ip,
		Ports:     ports,
		Labels:    dockerInfo.Labels,
		CreatedAt: time.Unix(dockerInfo.Created, 0).Format("2006-01-02 15:04:05"),
	}, nil
}

// IsReady checks if container is ready
func (dcm *DockerContainerManager) IsReady(ctx context.Context, containerName string) (bool, string, error) {
	// Check container status
	info, err := dcm.GetInfo(ctx, containerName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "No such container") {
			return false, "container does not exist", fmt.Errorf("failed to check container status: %w", err)
		}
		return false, err.Error(), err
	}

	status := info.Status
	if status != "running" {
		return false, fmt.Sprintf("container status: %s", status), nil
	}

	// Check health status (if health check is configured)
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Health.Status}}", containerName)
	output, err := cmd.Output()
	if err == nil {
		healthStatus := strings.TrimSpace(string(output))
		if healthStatus == "healthy" {
			return true, "container is running normally and health check passed", nil
		} else if healthStatus != "" && healthStatus != "<no value>" {
			return false, fmt.Sprintf("health check status: %s", healthStatus), nil
		}
	}

	// If no health check, consider ready as long as container is running
	return true, "container is running normally", nil
}

// GetEvents gets container events (Docker doesn't have direct event concept, returns log information)
func (dcm *DockerContainerManager) GetEvents(ctx context.Context, containerName string) ([]ContainerEvent, error) {
	// Docker doesn't have an event system like Kubernetes, here we return the last few lines of container logs as events
	logs, err := dcm.GetLogs(ctx, containerName, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	// Convert logs to events
	var events []ContainerEvent
	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			events = append(events, ContainerEvent{
				Type:      "Normal",
				Reason:    "Log",
				Message:   line,
				Timestamp: time.Now().Unix(),
			})
		}
	}

	return events, nil
}

// GetLogs gets container logs
func (dcm *DockerContainerManager) GetLogs(ctx context.Context, containerName string, lines int64) (string, error) {
	// Build docker logs command
	args := []string{"logs"}

	// Set line limit
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}

	// Add container name
	args = append(args, containerName)

	// Execute command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Docker container logs: %w", err)
	}

	return string(output), nil
}

// GetWarningEvents gets container warning events
func (dcm *DockerContainerManager) GetWarningEvents(ctx context.Context, containerName string) ([]ContainerEvent, error) {
	// Check if container has error status
	info, err := dcm.GetInfo(ctx, containerName)
	if err != nil {
		return nil, err
	}

	var events []ContainerEvent
	if info.Status != "running" {
		events = append(events, ContainerEvent{
			Type:      "Warning",
			Reason:    "ContainerNotRunning",
			Message:   fmt.Sprintf("container status abnormal: %s", info.Status),
			Timestamp: time.Now().Unix(),
		})
	}

	return events, nil
}

// getContainerIP gets container IP address
func (dcm *DockerContainerManager) getContainerIP(ctx context.Context, containerName string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// DockerServiceManager Docker service manager implementation (Docker doesn't have native service concept, using network aliases to simulate)
type DockerServiceManager struct {
	networkName string
}

// Create creates service (implemented through network aliases in Docker)
func (dsm *DockerServiceManager) Create(ctx context.Context, serviceName string, port int32, selector map[string]string) (*ServiceInfo, error) {
	// Docker doesn't have native service concept, here we create a custom network to simulate service discovery
	// In actual use, it might need to be combined with Docker Compose or other service discovery mechanisms

	// Check if network exists, create if not exists
	cmd := exec.CommandContext(ctx, "docker", "network", "inspect", serviceName)
	if err := cmd.Run(); err != nil {
		// Network doesn't exist, create network
		createCmd := exec.CommandContext(ctx, "docker", "network", "create", serviceName)
		if err := createCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to create Docker network: %w", err)
		}
	}

	return &ServiceInfo{
		Name:      serviceName,
		ClusterIP: "docker-network", // Docker network identifier
		Ports:     []int32{port},
		Labels:    selector,
	}, nil
}

// Delete deletes service
func (dsm *DockerServiceManager) Delete(ctx context.Context, serviceName string) error {
	cmd := exec.CommandContext(ctx, "docker", "network", "rm", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete Docker network: %w", err)
	}
	return nil
}

// Get gets service information
func (dsm *DockerServiceManager) Get(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "network", "inspect", serviceName)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get Docker network information: %w", err)
	}

	// Simple network information parsing
	return &ServiceInfo{
		Name:      serviceName,
		ClusterIP: "docker-network",
		Ports:     []int32{},               // Docker network itself doesn't contain port information
		Labels:    make(map[string]string), // Empty labels for Docker network
	}, nil
}

// Restart restarts service
func (dsm *DockerServiceManager) Restart(ctx context.Context, options ContainerCreateOptions) error {
	// Get existing service information
	existingService, err := dsm.Get(ctx, options.ServiceName)
	if err != nil {
		// If service doesn't exist, directly return error
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("service %s does not exist, cannot restart", options.ServiceName)
		}
		return fmt.Errorf("failed to get service information: %w", err)
	}

	// Delete existing service
	if err := dsm.Delete(ctx, options.ServiceName); err != nil {
		return fmt.Errorf("failed to delete existing service: %w", err)
	}

	// Wait for service to be completely deleted
	if err := dsm.waitForServiceDeletion(ctx, options.ServiceName); err != nil {
		return fmt.Errorf("failed to wait for service deletion completion: %w", err)
	}

	// Recreate service (use original port and labels)
	_, err = dsm.Create(ctx, options.ServiceName, options.Port, existingService.Labels)
	if err != nil {
		return fmt.Errorf("failed to recreate service %s: %w", options.ServiceName, err)
	}

	return nil
}

// waitForServiceDeletion waits for service to be completely deleted
func (dsm *DockerServiceManager) waitForServiceDeletion(ctx context.Context, serviceName string) error {
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
		_, err := dsm.Get(ctx, serviceName)
		if err != nil {
			// If get fails and is NotFound error, deletion is successful
			if strings.Contains(err.Error(), "not found") {
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
