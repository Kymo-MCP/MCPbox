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

// ContainerMonitorImpl 容器监控实现
type ContainerMonitorImpl struct {
	// instanceRepo 实例数据库操作
	instanceRepo *mysql.McpInstanceRepository

	// logger 日志记录器
	logger *zap.Logger

	// maxConcurrency 最大并发检查数
	maxConcurrency int
}

// NewContainerMonitor 创建新的容器监控器
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

// MonitorContainers 监控所有容器
func (cm *ContainerMonitorImpl) MonitorContainers(ctx context.Context) error {
	cm.logger.Info("开始执行全局容器监控任务")

	// 获取服务中托管实例
	instances, err := cm.instanceRepo.FindHostingInstances(ctx)
	if err != nil {
		cm.logger.Error("获取指定容器状态的MCP实例失败", zap.Error(err))
		return fmt.Errorf("获取指定容器状态的MCP实例失败: %w", err)
	}

	cm.logger.Info("获取到指定容器状态的MCP实例",
		zap.Int("count", len(instances)),
		zap.Strings("statuses", []string{string(model.ContainerStatusPending), string(model.ContainerStatusRunning)}))

	// 使用并发检查容器状态，最多同时检查 10 个
	semaphore := make(chan struct{}, cm.maxConcurrency)
	var wg sync.WaitGroup

	// 用于收集错误
	errorChan := make(chan error, len(instances))

	// 并发检查实例的容器状态
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *model.McpInstance) {
			defer wg.Done()

			// 获取信号量，控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			err := cm.CheckContainer(ctx, inst)
			if err != nil {
				cm.logger.Error("检查容器失败",
					zap.String("instance_id", inst.InstanceID),
					zap.String("container_name", inst.ContainerName),
					zap.Error(err))
				// 将错误发送到错误通道，但不阻塞
				select {
				case errorChan <- err:
				default:
				}
			}
		}(instance)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errorChan)

	// 收集并记录错误，但不中断整个监控流程
	errorCount := 0
	for err := range errorChan {
		errorCount++
		if errorCount == 1 {
			cm.logger.Warn("容器检查过程中出现错误", zap.Error(err))
		}
	}

	if errorCount > 0 {
		cm.logger.Warn("容器检查完成，部分实例检查失败",
			zap.Int("total_instances", len(instances)),
			zap.Int("failed_count", errorCount))
	}

	cm.logger.Info("全局容器监控任务执行完成")
	return nil
}

// CheckContainer 检查单个容器
func (cm *ContainerMonitorImpl) CheckContainer(ctx context.Context, instance *model.McpInstance) error {
	cm.logger.Debug("开始检查容器",
		zap.String("instance_id", instance.InstanceID))

	// 如果实例没有容器名称，说明还没有创建过容器，跳过
	if instance.ContainerName == "" {
		cm.logger.Debug("实例尚未创建容器，跳过检查",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// 获取创建参数
	containerCreateOptions := &container.ContainerCreateOptions{}
	if err := json.Unmarshal([]byte(instance.ContainerCreateOptions), containerCreateOptions); err != nil {
		cm.logger.Error("反序列化创建参数失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// 根据环境ID查询Kubernetes配置和命名空间
	environment, err := biz.GEnvironmentBiz.GetEnvironment(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("获取环境信息失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// 验证环境类型
	if environment.Environment != model.McpEnvironmentKubernetes {
		cm.logger.Error("环境类型错误，只支持 Kubernetes 环境",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// 获取当前时间戳（毫秒）
	currentTime := time.Now().UnixMilli()

	// 获取容器管理器
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("获取容器运行时失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	containerManager := entry.GetContainerManager()
	// 获取容器详细信息
	containerInfo, err := containerManager.GetInfo(ctx, instance.ContainerName)
	if err != nil {
		// 场景4: 容器不存在，重新创建并设置状态为创建中
		cm.logger.Warn("容器不存在，准备重新创建",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Error(err))

		return cm.recreateContainerWithStatus(ctx, instance, containerCreateOptions, model.ContainerStatusPending, "容器不存在，重新创建中")
	}

	// 解析容器创建时间（RFC3339格式）
	containerCreatedAt, err := time.Parse(time.RFC3339, containerInfo.CreatedAt)
	if err != nil {
		cm.logger.Warn("解析容器创建时间失败",
			zap.String("instance_id", instance.InstanceID),
			zap.String("created_at", containerInfo.CreatedAt),
			zap.Error(err))
		return err
	}
	containerCreatedAtMs := containerCreatedAt.UnixMilli()

	// 不等于运行中，检查启动超时，如果启动超时则清理容器
	if containerInfo.Status != "Running" {
		// 检查启动超时
		if instance.StartupTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.StartupTimeout {
				// 启动超时，清理容器和服务，更新状态
				startupDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("容器启动超时，清理资源",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_status", containerInfo.Status),
					zap.Int64("startup_duration_ms", startupDuration),
					zap.Int64("timeout_at_ms", instance.StartupTimeout))

				return cm.cleanupAndUpdateStatus(ctx, instance,
					fmt.Sprintf("容器启动超时，启动时长: %d毫秒，超时时间: %s", startupDuration,
						time.UnixMilli(containerCreatedAtMs).Format(time.RFC3339)))
			}
		}
	}

	// 检查容器是否就绪
	isReady, runInfo, err := containerManager.IsReady(ctx, instance.ContainerName)
	if err != nil {
		cm.logger.Error("检查容器就绪状态失败",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Error(err))
		return err
	}

	// 根据容器就绪状态进行处理
	if !isReady {
		// 检查启动超时
		if instance.StartupTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.StartupTimeout {
				// 启动超时，清理容器和服务，更新状态
				startupDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("容器启动超时，清理资源",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_status", containerInfo.Status),
					zap.Int64("startup_duration_ms", startupDuration),
					zap.Int64("timeout_at_ms", instance.StartupTimeout))

				return cm.cleanupAndUpdateStatus(ctx, instance,
					fmt.Sprintf("容器启动超时，启动时长: %d毫秒，超时时间: %s", startupDuration, time.UnixMilli(instance.StartupTimeout).Format(time.RFC3339)))
			}
		}

		// 检查运行超时（容器已启动但未就绪的情况）
		if containerInfo.Status == "Running" {
			if instance.RunningTimeout > 0 && (currentTime-containerCreatedAtMs) > instance.RunningTimeout {
				// 运行超时但未就绪，更新实例状态
				runningDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("容器运行中但未就绪，运行超时",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_name", instance.ContainerName),
					zap.Int64("running_duration_ms", runningDuration),
					zap.Int64("timeout_at_ms", instance.RunningTimeout),
					zap.String("run_info", runInfo))

				return cm.updateInstanceStatus(ctx, instance,
					model.ContainerStatusRunTimeoutStop,
					fmt.Sprintf("容器运行中但未就绪，运行超时，运行时长: %d毫秒，状态信息: %s", runningDuration, runInfo))
			}
		}

		// 实例容器状态是运行中，但是未就绪，更新实例状态
		if instance.ContainerStatus == model.ContainerStatusRunning {
			instance.ContainerIsReady = false
			instance.ContainerStatus = model.ContainerStatusRunningUnready
			instance.ContainerLastMessage = "容器运行中但未就绪"
			err := cm.instanceRepo.Update(ctx, instance)
			if err != nil {
				return fmt.Errorf("更新实例状态失败: %w", err)
			}
		}
		// 容器仍在启动中或运行中但未就绪，继续等待
		cm.logger.Debug("容器未就绪，继续等待",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_status", containerInfo.Status),
			zap.String("run_info", runInfo))

	} else {
		// 容器已就绪：检查运行超时
		if instance.RunningTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.RunningTimeout {
				// 运行超时，更新实例状态
				runningDuration := currentTime - containerCreatedAtMs
				message := fmt.Sprintf("容器运行超时，运行时长: %d毫秒，超时时间: %s", runningDuration, time.UnixMilli(instance.RunningTimeout).Format(time.RFC3339))
				cm.logger.Warn("容器运行超时",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_name", instance.ContainerName),
					zap.Int64("running_duration_ms", runningDuration),
					zap.Int64("timeout_at_ms", instance.RunningTimeout))

				return cm.updateInstanceStatus(ctx, instance, model.ContainerStatusRunTimeoutStop, message)
			}
		}

		// 容器正常运行且已就绪
		cm.logger.Debug("容器运行正常且已就绪",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName))

		// 确保实例状态为运行中
		if instance.ContainerStatus != model.ContainerStatusRunning {
			return cm.updateInstanceStatus(ctx, instance, model.ContainerStatusRunning, "容器运行正常且已就绪")
		}
	}

	return nil
}

// updateInstanceStatus 更新实例状态
func (cm *ContainerMonitorImpl) updateInstanceStatus(ctx context.Context, instance *model.McpInstance, containerStatus model.ContainerStatus, message string) error {

	instance.ContainerStatus = containerStatus
	instance.ContainerLastMessage = message

	err := cm.instanceRepo.Update(ctx, instance)
	if err != nil {
		cm.logger.Error("更新实例状态失败",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_status", string(containerStatus)),
			zap.Error(err))
		return fmt.Errorf("更新实例状态失败: %w", err)
	}

	cm.logger.Info("实例状态更新成功",
		zap.String("instance_id", instance.InstanceID),
		zap.String("container_status", string(containerStatus)),
		zap.String("message", message))

	return nil
}

// cleanupAndUpdateStatus 清理容器和服务，并更新状态为启动超时停止
func (cm *ContainerMonitorImpl) cleanupAndUpdateStatus(ctx context.Context, instance *model.McpInstance, message string) error {
	// 根据环境ID查询Kubernetes配置和命名空间
	environment, err := biz.GEnvironmentBiz.GetEnvironment(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("获取环境信息失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// 验证环境类型
	if environment.Environment != model.McpEnvironmentKubernetes {
		cm.logger.Error("环境类型错误，只支持 Kubernetes 环境",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// 获取容器管理器
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("获取容器运行时失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	containerManager := entry.GetContainerManager()
	serviceManager := entry.GetServiceManager()

	// 删除容器
	if instance.ContainerName != "" {
		err := containerManager.Delete(ctx, instance.ContainerName)
		if err != nil {
			cm.logger.Warn("删除容器失败",
				zap.String("instance_id", instance.InstanceID),
				zap.String("container_name", instance.ContainerName),
				zap.Error(err))
		}
	}

	// 删除服务
	if instance.ContainerServiceName != "" {
		err := serviceManager.Delete(ctx, instance.ContainerServiceName)
		if err != nil {
			cm.logger.Warn("删除服务失败",
				zap.String("instance_id", instance.InstanceID),
				zap.String("service_name", instance.ContainerServiceName),
				zap.Error(err))
		}
	}

	// 更新实例状态为启动超时停止
	return cm.updateInstanceStatus(ctx, instance,
		model.ContainerStatusInitTimeoutStop, message)
}

// recreateContainerWithStatus 重新创建容器并设置指定状态
func (cm *ContainerMonitorImpl) recreateContainerWithStatus(ctx context.Context, instance *model.McpInstance,
	options *container.ContainerCreateOptions, containerStatus model.ContainerStatus, message string) error {
	// 根据环境ID查询Kubernetes配置和命名空间
	environment, err := biz.GEnvironmentBiz.GetEnvironment(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("获取环境信息失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// 验证环境类型
	if environment.Environment != model.McpEnvironmentKubernetes {
		cm.logger.Error("环境类型错误，只支持 Kubernetes 环境",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// hosting 类型：查询容器状态
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		cm.logger.Error("获取容器运行时失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	containerManager := entry.GetContainerManager()
	serviceManager := entry.GetServiceManager()

	// 先删除旧容器（如果存在）
	if instance.ContainerName != "" {
		err = containerManager.Delete(ctx, instance.ContainerName)
		if err != nil {
			cm.logger.Warn("删除旧容器失败，继续创建新容器",
				zap.String("instance_id", instance.InstanceID),
				zap.String("container_name", instance.ContainerName),
				zap.Error(err))
		}
	}

	// 删除旧服务（如果存在）
	if instance.ContainerServiceName != "" {
		err = serviceManager.Delete(ctx, instance.ContainerServiceName)
		if err != nil {
			cm.logger.Warn("删除旧服务失败，继续创建新服务",
				zap.String("instance_id", instance.InstanceID),
				zap.String("service_name", instance.ContainerServiceName),
				zap.Error(err))
		}
	}

	// 创建新容器
	newContainerName, err := containerManager.Create(ctx, *options)
	if err != nil {
		cm.logger.Error("创建新容器失败",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return fmt.Errorf("创建新容器失败: %w", err)
	}

	// 创建新服务
	serviceName := instance.ContainerServiceName
	_, err = serviceManager.Create(ctx, serviceName, options.Port, options.Labels)
	if err != nil {
		cm.logger.Error("创建新服务失败",
			zap.String("instance_id", instance.InstanceID),
			zap.String("service_name", serviceName),
			zap.Error(err))

		// 删除已创建的容器
		if deleteErr := containerManager.Delete(ctx, newContainerName); deleteErr != nil {
			cm.logger.Error("删除容器失败",
				zap.String("container_name", newContainerName),
				zap.Error(deleteErr))
		}
		return fmt.Errorf("创建新服务失败: %w", err)
	}

	// 更新实例信息
	instance.ContainerName = newContainerName
	instance.ContainerServiceName = serviceName
	instance.ContainerStatus = containerStatus
	instance.ContainerLastMessage = message
	err = cm.instanceRepo.Update(ctx, instance)
	if err != nil {
		cm.logger.Error("更新实例容器信息失败",
			zap.String("instance_id", instance.InstanceID),
			zap.String("new_container_name", newContainerName),
			zap.Error(err))
		return fmt.Errorf("更新实例容器信息失败: %w", err)
	}

	cm.logger.Info("容器重建成功",
		zap.String("instance_id", instance.InstanceID),
		zap.String("new_container_name", newContainerName),
		zap.String("new_service_name", serviceName))

	return nil
}
