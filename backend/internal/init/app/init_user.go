package app

import (
	"context"
	"fmt"
	"qm-mcp-server/internal/authz/biz"
	"qm-mcp-server/internal/init/config"
	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
	"time"
)

func (a *App) createAdminUser() (*model.SysUser, error) {
	ctx := context.Background()
	userBiz := biz.NewUserBiz()
	initConfig := config.GetInitConfig()

	// Prepare admin user parameters
	now := time.Now()
	username := initConfig.Init.AdminUsername
	password := initConfig.Init.AdminPassword
	nickname := initConfig.Init.AdminNickname
	roleName := initConfig.Init.AdminRoleName
	level := initConfig.Init.AdminRoleLevel
	dataScope := initConfig.Init.AdminDataScope
	if roleName == "" {
		roleName = username
	}
	if level == 0 {
		level = 1
	}
	if dataScope == "" {
		dataScope = string(model.DataScopeAll)
	}
	isAdmin := true
	enabled := true

	adminUser := &model.SysUser{
		Username:   &username,
		NickName:   &nickname,
		IsAdmin:    isAdmin,
		Enabled:    &enabled,
		Source:     stringPtr("PLATFORM"),
		CreateTime: &now,
		UpdateTime: &now,
	}

	adminRole := &model.SysRole{
		Name:        roleName,
		Description: &roleName,
		Level:       &level,
		DataScope:   &dataScope,
		CreateTime:  &now,
		UpdateTime:  &now,
	}

	// Check if admin user already exists
	existingUser, err := mysql.SysUserRepo.FindByUsername(ctx, *adminUser.Username)
	if err == nil && existingUser != nil {
		fmt.Println("Admin user already exists, updating password...")
		// Update password
		err := userBiz.SetUserPassword(ctx, existingUser, password)
		if err != nil {
			return nil, fmt.Errorf("failed to update admin password: %v", err)
		}
		a.adminUser = existingUser
		return existingUser, nil
	}

	// Create admin role
	adminRole, err = createAdminRole(ctx, adminRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin role: %v", err)
	}

	// Create user
	err = userBiz.CreateUser(ctx, adminUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %v", err)
	}

	// Re-fetch user to ensure UserID is set correctly
	adminUser, err = mysql.SysUserRepo.FindByUsername(ctx, *adminUser.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created admin user: %v", err)
	}

	// Set user password
	err = userBiz.SetUserPassword(ctx, adminUser, password)
	if err != nil {
		return nil, fmt.Errorf("failed to set admin password: %v", err)
	}

	// Assign admin role to user
	err = userBiz.AssignRolesToUser(ctx, adminUser.UserID, []uint{adminRole.RoleID})
	if err != nil {
		return nil, fmt.Errorf("failed to assign admin role to user: %v", err)
	}

	fmt.Printf("Admin user created successfully with ID: %d\n", adminUser.UserID)
	a.adminUser = adminUser
	return adminUser, nil
}

func createAdminRole(ctx context.Context, adminRole *model.SysRole) (*model.SysRole, error) {
	// Check if admin role already exists
	existingRole, err := mysql.SysRoleRepo.FindByName(ctx, adminRole.Name)
	if err == nil && existingRole != nil {
		return existingRole, nil
	}

	// Create role
	err = mysql.SysRoleRepo.Create(ctx, adminRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin role: %v", err)
	}

	fmt.Printf("Admin role created successfully with ID: %d\n", adminRole.RoleID)
	return adminRole, nil
}

func stringPtr(s string) *string {
	return &s
}
