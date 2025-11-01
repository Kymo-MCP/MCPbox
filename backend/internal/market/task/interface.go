package task

import (
	"context"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
)

// TaskManager task manager interface
// Responsible for managing the lifecycle of global tasks
type TaskManager interface {
	// SetupGlobalTasks sets up global tasks
	// Initializes all required global tasks
	SetupGlobalTasks(ctx context.Context) error

	// StartMonitoring starts monitoring
	// Starts all monitoring tasks
	StartMonitoring(ctx context.Context) error

	// StopMonitoring stops monitoring
	// Stops all monitoring tasks
	StopMonitoring(ctx context.Context) error
}

// ContainerMonitor container monitoring interface
// Responsible for container status monitoring and management
type ContainerMonitor interface {
	// MonitorContainers monitors all containers
	// Checks the container status of all active instances
	MonitorContainers(ctx context.Context) error

	// CheckContainer checks a single container
	// Checks the container status of a specified instance and rebuilds if necessary
	CheckContainer(ctx context.Context, instance *model.McpInstance) error
}
