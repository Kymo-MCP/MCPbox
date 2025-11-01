package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"

	"go.uber.org/zap"
)

// ContainerMonitorImpl container monitoring implementation
type ContainerMonitorImpl struct {
	// instanceRepo instance database operations
	instanceRepo *mysql.McpInstanceRepository

	// logger logger
	logger *zap.Logger

	// maxConcurrency maximum concurrent check count
	maxConcurrency int
}

// NewContainerMonitor creates a new container monitor
func NewContainerMonitor(
	instanceRepo *mysql.McpInstanceRepository,
	logger *zap.Logger,
) ContainerMonitor {
	return &ContainerMonitorImpl{
		instanceRepo:   instanceRepo,
		logger:         logger,
		maxConcurrency: 10,
	}
}

// MonitorContainers monitors all containers
func (cm *ContainerMonitorImpl) MonitorContainers(ctx context.Context) error {
	cm.logger.Info("Starting global container monitoring task")

	// Get hosting instances in service
	instances, err := cm.instanceRepo.FindHostingInstances(ctx)
	if err != nil {
		cm.logger.Error("Failed to get MCP instances with specified container status", zap.Error(err))
		return fmt.Errorf("failed to get MCP instances with specified container status: %w", err)
	}

	cm.logger.Info("Retrieved MCP instances with specified container status",
		zap.Int("count", len(instances)),
		zap.Strings("statuses", []string{string(model.ContainerStatusPending), string(model.ContainerStatusRunning)}))

	// Use concurrent checking of container status, check at most 10 at the same time
	semaphore := make(chan struct{}, cm.maxConcurrency)
	var wg sync.WaitGroup

	// Used to collect errors
	errorChan := make(chan error, len(instances))

	// Concurrently check instance container status
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *model.McpInstance) {
			defer wg.Done()

			// Get semaphore to control concurrency
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			err := cm.CheckContainer(ctx, inst)
			if err != nil {
				cm.logger.Error("Container check failed",
					zap.String("instance_id", inst.InstanceID),
					zap.String("container_name", inst.ContainerName),
					zap.Error(err))
				// Send error to error channel, but don't block
				select {
				case errorChan <- err:
				default:
				}
			}
		}(instance)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Collect and log errors, but don't interrupt the entire monitoring process
	errorCount := 0
	for err := range errorChan {
		errorCount++
		if errorCount == 1 {
			cm.logger.Warn("Errors occurred during container checking", zap.Error(err))
		}
	}

	if errorCount > 0 {
		cm.logger.Warn("Container checking completed, some instance checks failed",
			zap.Int("total_instances", len(instances)),
			zap.Int("failed_count", errorCount))
	}

	cm.logger.Info("Global container monitoring task execution completed")
	return nil
}

// CheckContainer checks a single container
func (cm *ContainerMonitorImpl) CheckContainer(ctx context.Context, instance *model.McpInstance) error {
	cm.logger.Debug("Starting container check",
		zap.String("instance_id", instance.InstanceID))

	// If instance has no container name, it means no container has been created yet, skip
	if instance.ContainerName == "" {
		cm.logger.Debug("Instance has not created container yet, skipping check",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// Get creation parameters
	containerCreateOptions := &container.ContainerCreateOptions{}
	if err := json.Unmarshal([]byte(instance.ContainerCreateOptions), containerCreateOptions); err != nil {
		cm.logger.Error("Failed to deserialize creation parameters",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// Query Kubernetes configuration and namespace by environment ID
	environment, err := biz.GEnvironmentBiz.GetEnvironment(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("Failed to get environment information",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// Validate environment type
	if environment.Environment != model.McpEnvironmentKubernetes {
		cm.logger.Error("Environment type error, only Kubernetes environment is supported",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// Get current timestamp (milliseconds)
	currentTime := time.Now().UnixMilli()

	// Get container manager
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("Failed to get container runtime",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	containerManager := entry.GetContainerManager()
	// Get container detailed information
	containerInfo, err := containerManager.GetInfo(ctx, instance.ContainerName)
	if err != nil {
		// Scenario 4: Container does not exist, recreate and set status to creating
		cm.logger.Warn("Container does not exist, preparing to recreate",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Error(err))

		return cm.recreateContainerWithStatus(ctx, instance, containerCreateOptions, model.ContainerStatusPending, "Container does not exist, recreating")
	}

	// Parse container creation time (RFC3339 format)
	containerCreatedAt, err := time.Parse(time.RFC3339, containerInfo.CreatedAt)
	if err != nil {
		cm.logger.Warn("Failed to parse container creation time",
			zap.String("instance_id", instance.InstanceID),
			zap.String("created_at", containerInfo.CreatedAt),
			zap.Error(err))
		return err
	}
	containerCreatedAtMs := containerCreatedAt.UnixMilli()

	// Not equal to running, check startup timeout, if startup timeout then cleanup container
	if containerInfo.Status != "Running" {
		// Check startup timeout
		if instance.StartupTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.StartupTimeout {
				// Startup timeout, cleanup container and service, update status
				startupDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("Container startup timeout, cleaning up resources",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_status", containerInfo.Status),
					zap.Int64("startup_duration_ms", startupDuration),
					zap.Int64("timeout_at_ms", instance.StartupTimeout))

				return cm.cleanupAndUpdateStatus(ctx, instance,
					fmt.Sprintf("Container startup timeout, startup duration: %d milliseconds, timeout time: %s", startupDuration,
						time.UnixMilli(containerCreatedAtMs).Format(time.RFC3339)))
			}
		}
	}

	// Check if container is ready
	isReady, runInfo, err := containerManager.IsReady(ctx, instance.ContainerName)
	if err != nil {
		cm.logger.Error("Failed to check container ready status",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Error(err))
		return err
	}

	// Handle based on container ready status
	if !isReady {
		// Check startup timeout
		if instance.StartupTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.StartupTimeout {
				// Startup timeout, cleanup container and service, update status
				startupDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("Container startup timeout, cleaning up resources",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_status", containerInfo.Status),
					zap.Int64("startup_duration_ms", startupDuration),
					zap.Int64("timeout_at_ms", instance.StartupTimeout))

				return cm.cleanupAndUpdateStatus(ctx, instance,
					fmt.Sprintf("Container startup timeout, startup duration: %d milliseconds, timeout time: %s", startupDuration, time.UnixMilli(instance.StartupTimeout).Format(time.RFC3339)))
			}
		}

		// Check running timeout (container started but not ready)
		if containerInfo.Status == "Running" {
			if instance.RunningTimeout > 0 && (currentTime-containerCreatedAtMs) > instance.RunningTimeout {
				// Running timeout but not ready, update instance status
				runningDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("Container running but not ready, running timeout",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_name", instance.ContainerName),
					zap.Int64("running_duration_ms", runningDuration),
					zap.Int64("timeout_at_ms", instance.RunningTimeout),
					zap.String("run_info", runInfo))

				return cm.updateInstanceStatus(ctx, instance,
					model.ContainerStatusRunTimeoutStop,
					fmt.Sprintf("Container running but not ready, running timeout, running duration: %d milliseconds, status info: %s", runningDuration, runInfo))
			}
		}

		// Instance container status is running, but not ready, update instance status
		if instance.ContainerStatus == model.ContainerStatusRunning {
			instance.ContainerIsReady = false
			instance.ContainerStatus = model.ContainerStatusRunningUnready
			instance.ContainerLastMessage = "Container running but not ready"
			err := cm.instanceRepo.Update(ctx, instance)
			if err != nil {
				return fmt.Errorf("failed to update instance status: %w", err)
			}
		}
		// Container still starting or running but not ready, continue waiting
		cm.logger.Debug("Container not ready, continue waiting",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_status", containerInfo.Status),
			zap.String("run_info", runInfo))

	} else {
		// Container is ready: check running timeout
		if instance.RunningTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.RunningTimeout {
				// Running timeout, update instance status
				runningDuration := currentTime - containerCreatedAtMs
				message := fmt.Sprintf("Container running timeout, running duration: %d milliseconds, timeout time: %s", runningDuration, time.UnixMilli(instance.RunningTimeout).Format(time.RFC3339))
				cm.logger.Warn("Container running timeout",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_name", instance.ContainerName),
					zap.Int64("running_duration_ms", runningDuration),
					zap.Int64("timeout_at_ms", instance.RunningTimeout))

				return cm.updateInstanceStatus(ctx, instance, model.ContainerStatusRunTimeoutStop, message)
			}
		}

		// Container running normally and ready
		cm.logger.Debug("Container running normally and ready",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName))

		// Ensure instance status is running
		if instance.ContainerStatus != model.ContainerStatusRunning {
			return cm.updateInstanceStatus(ctx, instance, model.ContainerStatusRunning, "Container running normally and ready")
		}
	}

	return nil
}

// updateInstanceStatus updates instance status
func (cm *ContainerMonitorImpl) updateInstanceStatus(ctx context.Context, instance *model.McpInstance, containerStatus model.ContainerStatus, message string) error {

	instance.ContainerStatus = containerStatus
	instance.ContainerLastMessage = message

	err := cm.instanceRepo.Update(ctx, instance)
	if err != nil {
		cm.logger.Error("Failed to update instance status",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_status", string(containerStatus)),
			zap.Error(err))
		return fmt.Errorf("failed to update instance status: %w", err)
	}

	cm.logger.Info("Instance status updated successfully",
		zap.String("instance_id", instance.InstanceID),
		zap.String("container_status", string(containerStatus)),
		zap.String("message", message))

	return nil
}

// cleanupAndUpdateStatus cleans up container and service, and updates status to startup timeout stop
func (cm *ContainerMonitorImpl) cleanupAndUpdateStatus(ctx context.Context, instance *model.McpInstance, message string) error {
	// Query Kubernetes configuration and namespace by environment ID
	environment, err := biz.GEnvironmentBiz.GetEnvironment(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("Failed to get environment information",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// Validate environment type
	if environment.Environment != model.McpEnvironmentKubernetes {
		cm.logger.Error("Environment type error, only Kubernetes environment is supported",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// Get container manager
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("Failed to get container runtime",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	containerManager := entry.GetContainerManager()
	serviceManager := entry.GetServiceManager()

	// Delete container
	if instance.ContainerName != "" {
		err := containerManager.Delete(ctx, instance.ContainerName)
		if err != nil {
			cm.logger.Warn("Failed to delete container",
				zap.String("instance_id", instance.InstanceID),
				zap.String("container_name", instance.ContainerName),
				zap.Error(err))
		}
	}

	// Delete service
	if instance.ContainerServiceName != "" {
		err := serviceManager.Delete(ctx, instance.ContainerServiceName)
		if err != nil {
			cm.logger.Warn("Failed to delete service",
				zap.String("instance_id", instance.InstanceID),
				zap.String("service_name", instance.ContainerServiceName),
				zap.Error(err))
		}
	}

	// Update instance status to startup timeout stop
	return cm.updateInstanceStatus(ctx, instance,
		model.ContainerStatusInitTimeoutStop, message)
}

// recreateContainerWithStatus recreates container and sets specified status
func (cm *ContainerMonitorImpl) recreateContainerWithStatus(ctx context.Context, instance *model.McpInstance,
	options *container.ContainerCreateOptions, containerStatus model.ContainerStatus, message string) error {
	// Query Kubernetes configuration and namespace by environment ID
	environment, err := biz.GEnvironmentBiz.GetEnvironment(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("Failed to get environment information",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// Validate environment type
	if environment.Environment != model.McpEnvironmentKubernetes {
		cm.logger.Error("Environment type error, only Kubernetes environment is supported",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// hosting type: query container status
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("Failed to get container runtime",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	containerManager := entry.GetContainerManager()
	serviceManager := entry.GetServiceManager()

	// Delete old container first (if exists)
	if instance.ContainerName != "" {
		err = containerManager.Delete(ctx, instance.ContainerName)
		if err != nil {
			cm.logger.Warn("Failed to delete old container, continue creating new container",
				zap.String("instance_id", instance.InstanceID),
				zap.String("container_name", instance.ContainerName),
				zap.Error(err))
		}
	}

	// Delete old service (if exists)
	if instance.ContainerServiceName != "" {
		err = serviceManager.Delete(ctx, instance.ContainerServiceName)
		if err != nil {
			cm.logger.Warn("Failed to delete old service, continue creating new service",
				zap.String("instance_id", instance.InstanceID),
				zap.String("service_name", instance.ContainerServiceName),
				zap.Error(err))
		}
	}

	// Create new container
	newContainerName, err := containerManager.Create(ctx, *options)
	if err != nil {
		cm.logger.Error("Failed to create new container",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return fmt.Errorf("failed to create new container: %w", err)
	}

	// Create new service
	serviceName := instance.ContainerServiceName
	_, err = serviceManager.Create(ctx, serviceName, options.Port, options.Labels)
	if err != nil {
		cm.logger.Error("Failed to create new service",
			zap.String("instance_id", instance.InstanceID),
			zap.String("service_name", serviceName),
			zap.Error(err))

		// Delete created container
		if deleteErr := containerManager.Delete(ctx, newContainerName); deleteErr != nil {
			cm.logger.Error("Failed to delete container",
				zap.String("container_name", newContainerName),
				zap.Error(deleteErr))
		}
		return fmt.Errorf("failed to create new service: %w", err)
	}

	// Update instance information
	instance.ContainerName = newContainerName
	instance.ContainerServiceName = serviceName
	instance.ContainerStatus = containerStatus
	instance.ContainerLastMessage = message
	err = cm.instanceRepo.Update(ctx, instance)
	if err != nil {
		cm.logger.Error("Failed to update instance container information",
			zap.String("instance_id", instance.InstanceID),
			zap.String("new_container_name", newContainerName),
			zap.Error(err))
		return fmt.Errorf("failed to update instance container information: %w", err)
	}

	cm.logger.Info("Container recreated successfully",
		zap.String("instance_id", instance.InstanceID),
		zap.String("new_container_name", newContainerName),
		zap.String("new_service_name", serviceName))

	return nil
}
