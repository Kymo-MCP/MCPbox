package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "github.com/kymo-mcp/mcpcan/api/market/dashboard"
	instancepb "github.com/kymo-mcp/mcpcan/api/market/instance"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// DashboardService dashboard service
type DashboardService struct {
	pb.UnimplementedDashboardServiceServer

	instanceBiz    *biz.InstanceBiz
	environmentBiz *biz.EnvironmentBiz
	ctx            context.Context
}

// NewDashboardService create dashboard service
func NewDashboardService(ctx context.Context) *DashboardService {
	return &DashboardService{
		instanceBiz:    biz.GInstanceBiz,
		environmentBiz: biz.GEnvironmentBiz,
		ctx:            ctx,
	}
}

// Statistical get statistical data
func (s *DashboardService) Statistical(ctx context.Context, req *pb.StatisticalRequest) (*pb.StatisticalResponse, error) {
	// Get all instances
	instances, err := s.instanceBiz.ListInstance(1, 10000, nil, "", "")
	if err != nil {
		logger.Error("Failed to get all instances", zap.Error(err))
		return nil, err
	}

	// Get all environments
	environments, err := s.environmentBiz.ListEnvironments(ctx)
	if err != nil {
		logger.Error("Failed to get all environments", zap.Error(err))
		return nil, err
	}

	// Count statistics
	var totalInstances, activeInstances, inactiveInstances int32
	var proxyInstances, directInstances, hostingInstances int32

	for _, instance := range instances.List {
		totalInstances++

		// Count by status
		if instance.Status == string(model.InstanceStatusActive) {
			activeInstances++
		} else {
			inactiveInstances++
		}

		// Count by access type
		switch instance.AccessType {
		case instancepb.AccessType_PROXY:
			proxyInstances++
		case instancepb.AccessType_DIRECT:
			directInstances++
		case instancepb.AccessType_HOSTING:
			hostingInstances++
		}
	}

	return &pb.StatisticalResponse{
		TotalInstances:    totalInstances,
		ActiveInstances:   activeInstances,
		InactiveInstances: inactiveInstances,
		ProxyInstances:    proxyInstances,
		DirectInstances:   directInstances,
		HostingInstances:  hostingInstances,
		TotalEnvironments: int32(len(environments)),
	}, nil
}

// AvailableCases returns the top 4 templates ordered by creation time
func (s *DashboardService) AvailableCases(ctx context.Context, req *pb.AvailableCasesRequest) (*pb.AvailableCasesResponse, error) {
	// Query top 4 templates ordered by creation time
	templates, _, err := biz.GTemplateBiz.GetTemplatesWithPagination(ctx, 1, 4, nil, "created_at", "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %v", err)
	}

	// Convert to response format
	cases := make([]*pb.CaseInfo, 0, len(templates))
	for _, template := range templates {
		caseInfo := &pb.CaseInfo{
			TemplateId:  int32(template.ID),
			Name:        template.Name,
			Description: template.Notes,
			CreatedAt:   template.CreatedAt.Unix(),
			IconPath:    template.IconPath,
		}
		cases = append(cases, caseInfo)
	}

	return &pb.AvailableCasesResponse{
		Cases: cases,
	}, nil
}

func (s *DashboardService) StatisticalHandler(c *gin.Context) {
	req := &pb.StatisticalRequest{}
	resp, err := s.Statistical(c.Request.Context(), req)
	if err != nil {
		common.GinError(c, i18n.CodeInternalError, err.Error())
		return
	}
	common.GinSuccess(c, resp)
}

func (s *DashboardService) AvailableCasesHandler(c *gin.Context) {
	req := &pb.AvailableCasesRequest{}
	resp, err := s.AvailableCases(c.Request.Context(), req)
	if err != nil {
		common.GinError(c, i18n.CodeInternalError, err.Error())
		return
	}
	common.GinSuccess(c, resp)
}
