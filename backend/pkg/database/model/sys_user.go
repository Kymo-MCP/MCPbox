package model

import (
	"fmt"
	"strings"
	"time"
)

// UserSource 用户来源枚举
type UserSource string

const (
	UserSourcePlatform UserSource = "PLATFORM" // 自建
	UserSourceFeishu   UserSource = "FEISHU"   // 飞书
)

// SysUser 系统用户模型
type SysUser struct {
	UserID             uint       `gorm:"column:user_id;primaryKey;autoIncrement;comment:ID" json:"userId"`
	DeptID             *uint      `gorm:"column:dept_id;comment:部门名称" json:"deptId"`
	Username           *string    `gorm:"column:username;size:180;uniqueIndex:uniq_username;comment:用户名" json:"username"`
	NickName           *string    `gorm:"column:nick_name;size:255;comment:昵称" json:"nickName"`
	Gender             *string    `gorm:"column:gender;size:2;comment:性别" json:"gender"`
	Phone              *string    `gorm:"column:phone;size:255;comment:手机号码" json:"phone"`
	Email              *string    `gorm:"column:email;size:180;uniqueIndex:uniq_email;comment:邮箱" json:"email"`
	AvatarName         *string    `gorm:"column:avatar_name;size:255;comment:头像地址" json:"avatarName"`
	AvatarPath         *string    `gorm:"column:avatar_path;size:255;comment:头像真实路径" json:"avatarPath"`
	Password           *string    `gorm:"column:password;size:255;comment:密码" json:"password"`
	Salt               *string    `gorm:"column:salt;size:255;comment:密码盐" json:"salt"`
	IsAdmin            bool       `gorm:"column:is_admin;default:false;comment:是否为admin账号" json:"isAdmin"`
	Enabled            *bool      `gorm:"column:enabled;index:inx_enabled;comment:状态：1启用、0禁用" json:"enabled"`
	CreateBy           *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy           *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	PwdResetTime       *time.Time `gorm:"column:pwd_reset_time;comment:修改密码的时间" json:"pwdResetTime"`
	CreateTime         *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime         *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	EnterpriseWechatID *string    `gorm:"column:enterprise_wechat_id;size:255;comment:企业微信ID" json:"enterpriseWechatId"`
	CreateQAgent       bool       `gorm:"column:create_qagent;default:false;comment:是否创建QAgent账号" json:"createQAgent"`
	DingTalkID         *string    `gorm:"column:ding_talk_id;size:128;comment:钉钉ID" json:"dingTalkId"`
	FeishuID           *string    `gorm:"column:feishu_id;size:128;comment:飞书ID" json:"feishuId"`
	Source             *string    `gorm:"column:source;size:32;comment:来源 PLATFORM：自建，FEISHU:飞书" json:"source"`
	ThirdPartyOpenID   *string    `gorm:"column:third_party_open_id;size:128;comment:第三方平台唯一id 飞书：openId" json:"thirdPartyOpenId"`
	ThirdPartyUnionID  *string    `gorm:"column:third_party_union_id;size:128;comment:第三方平台唯一id[跨应用] 飞书：union_id" json:"thirdPartyUnionId"`
	CorpID             *string    `gorm:"column:corp_id;size:128;comment:来源中的corpId 飞书：corpId 钉钉：corpId" json:"corpId"`
}

// TableName 返回表名
func (SysUser) TableName() string {
	return "sys_user"
}

// PrepareForCreate 创建前的准备工作
func (u *SysUser) PrepareForCreate() error {
	now := time.Now()
	u.CreateTime = &now
	u.UpdateTime = &now
	return nil
}

// PrepareForUpdate 更新前的准备工作
func (u *SysUser) PrepareForUpdate() error {
	now := time.Now()
	u.UpdateTime = &now
	return nil
}

// ValidateForCreate 创建前验证
func (u *SysUser) ValidateForCreate() error {
	if u.Username == nil || strings.TrimSpace(*u.Username) == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if u.Email != nil && *u.Email != "" {
		if !isValidEmail(*u.Email) {
			return fmt.Errorf("邮箱格式不正确")
		}
	}
	if u.Source != nil && !isValidUserSource(*u.Source) {
		return fmt.Errorf("用户来源不正确，必须是 PLATFORM 或 FEISHU")
	}
	return nil
}

// ValidateForUpdate 更新前验证
func (u *SysUser) ValidateForUpdate() error {
	if u.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if u.Email != nil && *u.Email != "" {
		if !isValidEmail(*u.Email) {
			return fmt.Errorf("邮箱格式不正确")
		}
	}
	if u.Source != nil && !isValidUserSource(*u.Source) {
		return fmt.Errorf("用户来源不正确，必须是 PLATFORM 或 FEISHU")
	}
	return nil
}

// Clone 克隆对象
func (u *SysUser) Clone() *SysUser {
	if u == nil {
		return nil
	}
	clone := &SysUser{
		UserID:       u.UserID,
		IsAdmin:      u.IsAdmin,
		CreateQAgent: u.CreateQAgent,
	}

	// 复制指针字段
	if u.DeptID != nil {
		deptID := *u.DeptID
		clone.DeptID = &deptID
	}
	if u.Username != nil {
		username := *u.Username
		clone.Username = &username
	}
	if u.NickName != nil {
		nickName := *u.NickName
		clone.NickName = &nickName
	}
	if u.Gender != nil {
		gender := *u.Gender
		clone.Gender = &gender
	}
	if u.Phone != nil {
		phone := *u.Phone
		clone.Phone = &phone
	}
	if u.Email != nil {
		email := *u.Email
		clone.Email = &email
	}
	if u.AvatarName != nil {
		avatarName := *u.AvatarName
		clone.AvatarName = &avatarName
	}
	if u.AvatarPath != nil {
		avatarPath := *u.AvatarPath
		clone.AvatarPath = &avatarPath
	}
	if u.Password != nil {
		password := *u.Password
		clone.Password = &password
	}
	if u.Enabled != nil {
		enabled := *u.Enabled
		clone.Enabled = &enabled
	}
	if u.CreateBy != nil {
		createBy := *u.CreateBy
		clone.CreateBy = &createBy
	}
	if u.UpdateBy != nil {
		updateBy := *u.UpdateBy
		clone.UpdateBy = &updateBy
	}
	if u.PwdResetTime != nil {
		pwdResetTime := *u.PwdResetTime
		clone.PwdResetTime = &pwdResetTime
	}
	if u.CreateTime != nil {
		createTime := *u.CreateTime
		clone.CreateTime = &createTime
	}
	if u.UpdateTime != nil {
		updateTime := *u.UpdateTime
		clone.UpdateTime = &updateTime
	}
	if u.EnterpriseWechatID != nil {
		enterpriseWechatID := *u.EnterpriseWechatID
		clone.EnterpriseWechatID = &enterpriseWechatID
	}
	if u.DingTalkID != nil {
		dingTalkID := *u.DingTalkID
		clone.DingTalkID = &dingTalkID
	}
	if u.FeishuID != nil {
		feishuID := *u.FeishuID
		clone.FeishuID = &feishuID
	}
	if u.Source != nil {
		source := *u.Source
		clone.Source = &source
	}
	if u.ThirdPartyOpenID != nil {
		thirdPartyOpenID := *u.ThirdPartyOpenID
		clone.ThirdPartyOpenID = &thirdPartyOpenID
	}
	if u.ThirdPartyUnionID != nil {
		thirdPartyUnionID := *u.ThirdPartyUnionID
		clone.ThirdPartyUnionID = &thirdPartyUnionID
	}
	if u.CorpID != nil {
		corpID := *u.CorpID
		clone.CorpID = &corpID
	}

	return clone
}

// 用户来源相关方法

// IsThirdParty 是否为第三方用户
func (u *SysUser) IsThirdParty() bool {
	return u.Source != nil && *u.Source != string(UserSourcePlatform)
}

// IsFeishuUser 是否为飞书用户
func (u *SysUser) IsFeishuUser() bool {
	return u.Source != nil && *u.Source == string(UserSourceFeishu)
}

// IsPlatformUser 是否为平台自建用户
func (u *SysUser) IsPlatformUser() bool {
	return u.Source == nil || *u.Source == string(UserSourcePlatform)
}

// 状态相关方法

// IsEnabled 是否启用
func (u *SysUser) IsEnabled() bool {
	return u.Enabled != nil && *u.Enabled
}

// IsDisabled 是否禁用
func (u *SysUser) IsDisabled() bool {
	return u.Enabled != nil && !*u.Enabled
}

// 部门相关方法

// HasDept 是否有部门
func (u *SysUser) HasDept() bool {
	return u.DeptID != nil && *u.DeptID > 0
}

// 第三方平台相关方法

// HasFeishuID 是否有飞书ID
func (u *SysUser) HasFeishuID() bool {
	return u.FeishuID != nil && strings.TrimSpace(*u.FeishuID) != ""
}

// HasDingTalkID 是否有钉钉ID
func (u *SysUser) HasDingTalkID() bool {
	return u.DingTalkID != nil && strings.TrimSpace(*u.DingTalkID) != ""
}

// HasEnterpriseWechatID 是否有企业微信ID
func (u *SysUser) HasEnterpriseWechatID() bool {
	return u.EnterpriseWechatID != nil && strings.TrimSpace(*u.EnterpriseWechatID) != ""
}

// Getter 方法

// GetUsername 获取用户名
func (u *SysUser) GetUsername() string {
	if u.Username == nil {
		return ""
	}
	return *u.Username
}

// GetNickName 获取昵称
func (u *SysUser) GetNickName() string {
	if u.NickName == nil {
		return ""
	}
	return *u.NickName
}

// GetEmail 获取邮箱
func (u *SysUser) GetEmail() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

// GetPhone 获取手机号
func (u *SysUser) GetPhone() string {
	if u.Phone == nil {
		return ""
	}
	return *u.Phone
}

// GetSource 获取用户来源
func (u *SysUser) GetSource() string {
	if u.Source == nil {
		return string(UserSourcePlatform)
	}
	return *u.Source
}

// GetDeptID 获取部门ID
func (u *SysUser) GetDeptID() uint {
	if u.DeptID == nil {
		return 0
	}
	return *u.DeptID
}

// Setter 方法

// SetUsername 设置用户名
func (u *SysUser) SetUsername(username string) {
	u.Username = &username
}

// SetNickName 设置昵称
func (u *SysUser) SetNickName(nickName string) {
	u.NickName = &nickName
}

// SetEmail 设置邮箱
func (u *SysUser) SetEmail(email string) {
	u.Email = &email
}

// SetPhone 设置手机号
func (u *SysUser) SetPhone(phone string) {
	u.Phone = &phone
}

// SetPassword 设置密码
func (u *SysUser) SetPassword(password string) {
	u.Password = &password
}

// SetEnabled 设置启用状态
func (u *SysUser) SetEnabled(enabled bool) {
	u.Enabled = &enabled
}

// SetDeptID 设置部门ID
func (u *SysUser) SetDeptID(deptID uint) {
	u.DeptID = &deptID
}

// SetSource 设置用户来源
func (u *SysUser) SetSource(source string) {
	u.Source = &source
}

// SetFeishuID 设置飞书ID
func (u *SysUser) SetFeishuID(feishuID string) {
	u.FeishuID = &feishuID
}

// SetCorpID 设置企业ID
func (u *SysUser) SetCorpID(corpID string) {
	u.CorpID = &corpID
}

// SetCreateBy 设置创建者
func (u *SysUser) SetCreateBy(createBy string) {
	u.CreateBy = &createBy
}

// SetUpdateBy 设置更新者
func (u *SysUser) SetUpdateBy(updateBy string) {
	u.UpdateBy = &updateBy
}

// 辅助函数

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// isValidUserSource 验证用户来源
func isValidUserSource(source string) bool {
	return source == string(UserSourcePlatform) || source == string(UserSourceFeishu)
}
