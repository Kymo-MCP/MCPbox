package i18n

// 错误码定义
const (
	// 成功
	CodeSuccess = 0

	// 通用错误 (1000-1999)
	CodeBadRequest            = 1000
	CodeUnauthorized          = 1001
	CodeForbidden             = 1002
	CodeNotFound              = 1003
	CodeMethodNotAllowed      = 1004
	CodeRequestTimeout        = 1005
	CodeTooManyRequests       = 1006
	CodeInternalError         = 1007
	CodeNotImplemented        = 1008
	CodeServiceUnavailable    = 1009
	CodeGatewayTimeout        = 1010
	CodeInvalidPathParameters = 1011

	// 认证相关错误 (2000-2999)
	CodeInvalidToken       = 2000
	CodeTokenExpired       = 2001
	CodeMissingToken       = 2002
	CodeInvalidCredentials = 2003
	CodeUserNotFound       = 2004
	CodePasswordIncorrect  = 2005
	CodeUserDisabled       = 2006
	CodeAccountLocked      = 2007

	// 授权相关错误 (3000-3999)
	CodeInsufficientPermissions = 3000
	CodeAccessDenied            = 3001
	CodeRoleRequired            = 3002
	CodePermissionRequired      = 3003

	// 请求签名相关错误 (4000-4999)
	CodeInvalidSignature = 4000
	CodeSignatureExpired = 4001
	CodeMissingSignature = 4002
	CodeReplayAttack     = 4003
	CodeTimestampInvalid = 4004
	CodeKeyNotFound      = 4005
	CodeKeyExpired       = 4006

	// 业务逻辑错误 (5000-5999)
	CodeBusinessError       = 5000
	CodeDataValidation      = 5001
	CodeDuplicateEntry      = 5002
	CodeForeignKeyViolation = 5003
	CodeDataConflict        = 5004
	CodeResourceExhausted   = 5005

	// 系统错误 (6000-6999)
	CodeDatabaseError      = 6000
	CodeNetworkError       = 6001
	CodeFileSystemError    = 6002
	CodeConfigurationError = 6003
	CodeTimeoutError       = 6004
	CodeDependencyError    = 6005

	// 参数化错误码 (7000-7999)
	CodeParameterRequired     = 7000
	CodeParameterInvalid      = 7001
	CodeResourceNotFound      = 7002
	CodeResourceAlreadyExists = 7003
	CodeOperationFailed       = 7004
	CodeServiceError          = 7005
	CodeConnectionFailed      = 7006
	CodeFileOperationFailed   = 7007
	CodeParseError            = 7008
	CodeValidationFailed      = 7009

	// 任务调度相关错误 (8000-8099)
	CodeTaskCannotBeEmpty         = 8000
	CodeTaskIDCannotBeEmpty       = 8001
	CodeTaskIDNotExists           = 8002
	CodeUnsupportedTaskType       = 8003
	CodeTaskAlreadyExists         = 8004
	CodeTaskSaveFailure           = 8005
	CodeTaskDeleteFailure         = 8006
	CodeTaskExecutionFailure      = 8007
	CodeInvalidCronExpression     = 8008
	CodeSchedulerAlreadyRunning   = 8009
	CodeSchedulerNotRunning       = 8010
	CodeTaskFunctionNotSet        = 8011
	CodeTaskFunctionNotExists     = 8012
	CodeTaskFunctionAlreadyExists = 8013
	CodeExecutionTimeTooEarly     = 8014

	// 用户认证相关错误 (8100-8199)
	CodeUsernameOrPasswordIncorrect = 8100
	CodeUserDisabledError           = 8101
	CodeLoginFailure                = 8102
	CodeLogoutFailure               = 8103
	CodeRefreshTokenInvalid         = 8104
	CodeRefreshTokenExpired         = 8105
	CodeRefreshFailure              = 8106
	CodeGenerateTokenFailure        = 8107
	CodeUserNotFoundError           = 8108
	CodeOldPasswordIncorrect        = 8109
	CodePasswordHashFailure         = 8110
	CodeUpdatePasswordFailure       = 8111
	CodeGenerateSaltFailure         = 8112
	CodeUserIDInvalid               = 8113
	CodeOldPasswordEmpty            = 8114
	CodeNewPasswordEmpty            = 8115
	CodeConfirmPasswordEmpty        = 8116
	CodePasswordMismatch            = 8117
	CodePasswordTooWeak             = 8118
	CodePasswordSameAsOld           = 8119

	// 用户管理相关错误 (8200-8299)
	CodeUsernameAlreadyExists   = 8200
	CodeEmailAlreadyExists      = 8201
	CodeCreateUserFailure       = 8202
	CodeUpdateUserFailure       = 8203
	CodeDeleteUserFailure       = 8204
	CodeGetUserFailure          = 8205
	CodeUserRoleAssignFailure   = 8206
	CodeUserRoleDeleteFailure   = 8207
	CodeUserRoleQueryFailure    = 8208
	CodeUserRoleAlreadyExists   = 8209
	CodeBatchUserRoleAddFailure = 8210
	CodeUserPasswordSetFailure  = 8211

	// 角色管理相关错误 (8300-8399)
	CodeRoleDataValidationFailure = 8300
	CodeRoleCreateFailure         = 8301
	CodeRoleUpdateFailure         = 8302
	CodeRoleDeleteFailure         = 8303
	CodeRoleQueryFailure          = 8304

	// 容器相关错误 (8400-8499)
	CodeContainerRuntimeNotInitialized = 8400
	CodeContainerCreateFailure         = 8401
	CodeContainerDeleteFailure         = 8402
	CodeContainerRestartFailure        = 8403
	CodeContainerStatusCheckFailure    = 8404
	CodeContainerLogGetFailure         = 8405
	CodeServiceCreateFailure           = 8406
	CodeServiceDeleteFailure           = 8407
	CodeInstanceNotExists              = 8408
	CodeInstanceContainerNotExists     = 8409
	CodeInstanceEnvironmentIDNotExists = 8410
	CodeGetEnvironmentRuntimeFailure   = 8411
	CodeUnsupportedEnvironmentType     = 8412
	CodeImageAddressRequired           = 8413
	CodePortRequired                   = 8414
	CodeStartupCommandRequired         = 8415
	CodeContainerNotReady              = 8416
	CodeSetReplicasFailure             = 8417
	CodeUpdateInstanceStatusFailure    = 8418
	CodeServiceNoRestartNeeded         = 8419
	CodeDockerEnvironmentNotSupported  = 8420

	// Kubernetes 相关错误 (8500-8599)
	CodeK8sClientInitFailure       = 8500
	CodeDeploymentCreateFailure    = 8501
	CodeDeploymentDeleteFailure    = 8502
	CodeDeploymentGetFailure       = 8503
	CodeDeploymentUpdateFailure    = 8504
	CodePodListGetFailure          = 8505
	CodePodLogGetFailure           = 8506
	CodePodNotFound                = 8507
	CodeServiceGetFailure          = 8508
	CodePVCCreateFailure           = 8509
	CodePVCDeleteFailure           = 8510
	CodePVCGetFailure              = 8511
	CodePVCNotBound                = 8512
	CodePVNotFound                 = 8513
	CodeNodeInfoExtractFailure     = 8514
	CodeNamespaceListGetFailure    = 8515
	CodeNodeListGetFailure         = 8516
	CodeStorageClassListGetFailure = 8517
	CodeVolumeTypeNotSupported     = 8518
	CodeHostPathCannotBeEmpty      = 8519
	CodeNodeNameCannotBeEmpty      = 8520
	CodePVCNameCannotBeEmpty       = 8521
	CodeConfigMapNameCannotBeEmpty = 8522
	CodeMountPathCannotBeEmpty     = 8523

	// 加密相关错误 (8600-8699)
	CodeRSAKeyGenerateFailure     = 8600
	CodePrivateKeyEncodeFailure   = 8601
	CodePublicKeyEncodeFailure    = 8602
	CodePublicKeyParseFailure     = 8603
	CodeRSAEncryptFailure         = 8604
	CodePrivateKeyParseFailure    = 8605
	CodeRSADecryptFailure         = 8606
	CodePasswordDataParseFailure  = 8607
	CodePasswordExpired           = 8608
	CodeInvalidPEMFormat          = 8609
	CodeNotRSAKey                 = 8610
	CodeBase64DecodeFailure       = 8611
	CodeRandomSaltGenerateFailure = 8612
	CodeBcryptEncryptFailure      = 8613
	CodeGetPrivateKeyFailure      = 8614

	// 安全相关错误 (8700-8799)
	CodeTimestampMissing            = 8700
	CodeTimestampFormatError        = 8701
	CodeRequestExpired              = 8702
	CodeRequestTooEarly             = 8703
	CodeSignatureMissing            = 8704
	CodeSignatureStringBuildFailure = 8705
	CodeSignatureMismatch           = 8706

	// 环境相关错误 (8800-8899)
	CodeEnvironmentNameCannotBeEmpty     = 8800
	CodeEnvironmentTypeCannotBeEmpty     = 8801
	CodeEnvironmentAlreadyExists         = 8802
	CodeEnvironmentNotFound              = 8803
	CodeEnvironmentCreateFailure         = 8804
	CodeEnvironmentUpdateFailure         = 8805
	CodeEnvironmentDeleteFailure         = 8806
	CodeEnvironmentTestFailure           = 8807
	CodeEnvironmentConfigInvalid         = 8808
	CodeEnvironmentConfigRequired        = 8809
	CodeEnvironmentTypeInvalid           = 8810
	CodeEnvironmentNameTooLong           = 8811
	CodeEnvironmentDescTooLong           = 8812
	CodeEnvironmentInUse                 = 8813
	CodeEnvironmentNotExists             = 8814
	CodeOnlyK8sSupportNamespace          = 8816
	CodeKubeconfigMissingField           = 8817
	CodeKubeconfigFormatError            = 8818
	CodeRuntimeTypeError                 = 8819
	CodeGetK8sEntryFailure               = 8820
	CodeK8sEntryAssertionFailure         = 8821
	CodeEnvironmentNotK8s                = 8822
	CodeDockerConnectionSuccess          = 8823 // Docker连接测试成功
	CodeKubeconfigParseFailure           = 8824 // kubeconfig解析失败
	CodeKubeconfigYamlConversionFailure  = 8825 // kubeconfig转换为YAML失败
	CodeKubeconfigConversionFailure      = 8826 // kubeconfig配置转换失败
	CodeListNamespacesFailure            = 8828 // 获取命名空间列表失败
	CodeGetRuntimeEntryFailure           = 8829 // 获取环境运行时入口失败
	CodeContainerCreateSuccess           = 8833 // 容器创建成功
	CodeContainerDeleteSuccess           = 8837 // 删除容器成功
	CodeServiceDeleteSuccess             = 8839 // 删除服务成功
	CodeInstanceNotHostingMode           = 8840 // 实例不存在或不是托管模式
	CodeContainerReadyCheckFailure       = 8841 // 检查容器就绪状态失败
	CodeGetContainerWarningEventsFailure = 8843 // 获取容器警告事件失败
	CodeServiceStatusAbnormal            = 8844 // 服务状态异常
	CodeUpdateInstanceFailure            = 8845 // 更新实例失败
	CodeDeleteContainerFailure           = 8850 // 删除容器失败
	CodeContainerScaledToZero            = 8851 // 容器已缩放至0副本
	CodeGetContainerLogsFailure          = 8852 // 获取容器日志失败
	CodeParseContainerOptionsFailure     = 8853 // 解析容器创建选项失败
	CodeMissingContainerOptions          = 8854 // 实例缺少容器创建选项
	CodeRestartContainerFailure          = 8855 // 重启容器失败
	CodeRestartContainerSuccess          = 8856 // 容器重启成功
	CodeGetEnvironmentInfoFailure        = 8857 // 获取环境信息失败
	CodeGetK8sRuntimeEntryFailure        = 8858 // 获取Kubernetes运行时入口失败
	CodeFailedToFindCodePackage          = 8861 // 查找代码包失败
	CodeFailedToGenerateDownloadZip      = 8862 // 生成下载ZIP包失败

	// 实例相关错误 (8900-8999)
	CodeInstanceNameAlreadyExists  = 8900
	CodeInstanceQueryFailure       = 8901
	CodeInstanceUpdateFailure      = 8902
	CodeInstanceDeleteFailure      = 8903
	CodeInstanceDisableFailure     = 8904
	CodeInstanceRestartFailure     = 8905
	CodeInstanceConfigBuildFailure = 8906
	CodeGetContainerStatusFailure  = 8907
	CodeGetTargetConfigFailure     = 8908
	CodeUnsupportedAccessType      = 8909
	CodeAccessTypeConvertFailure   = 8910
	CodeMCPProtocolConvertFailure  = 8911

	// 数据库相关错误 (9000-9099)
	CodeEncryptionKeyNotExists   = 9000
	CodeGetEncryptionKeyFailure  = 9001
	CodeGetActiveKeyFailure      = 9002
	CodeCreateKeyFailure         = 9003
	CodeUpdateKeyStatusFailure   = 9004
	CodeDeleteExpiredKeysFailure = 9005
	CodeGetClientKeyListFailure  = 9006
	CodeCountKeysFailure         = 9007

	// 通用参数验证错误 (9100-9199)
	CodeUserIDCannotBeEmpty               = 9100
	CodeRoleIDCannotBeEmpty               = 9101
	CodeDeptIDCannotBeEmpty               = 9102
	CodeUsernameCannotBeEmpty             = 9103
	CodeEmailCannotBeEmpty                = 9104
	CodeEmailFormatIncorrect              = 9105
	CodeUserSourceIncorrect               = 9106
	CodeRoleNameCannotBeEmpty             = 9107
	CodeRoleNameTooLong                   = 9108
	CodeInvalidDataScope                  = 9109
	CodeRoleLevelCannotBeNegative         = 9110
	CodeDeptNameCannotBeEmpty             = 9111
	CodeInvalidDeptSource                 = 9112
	CodeFeishuDeptMustProvideEnterpriseID = 9113
	CodeFeishuDeptMustProvideOpenDeptID   = 9114
	CodeImageNameCannotBeEmpty            = 9115
	CodePodNameCannotBeEmpty              = 9116
	CodeAppNameCannotBeEmpty              = 9117
	CodeStorageSizeCannotBeEmpty          = 9118
	CodeInvalidAccessMode                 = 9119
	CodeConfigParameterCannotBeEmpty      = 9120
)
