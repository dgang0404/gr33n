package commontypes

// ── User & Farm roles ─────────────────────────────────────────────────────────

type UserRoleEnum string
const (
	UserRoleAdmin       UserRoleEnum = "admin"
	UserRoleFarmManager UserRoleEnum = "farm_manager"
	UserRoleOperator    UserRoleEnum = "operator"
	UserRoleViewer      UserRoleEnum = "viewer"
	UserRoleSystemAdmin UserRoleEnum = "gr33n_system_admin"
)

type FarmMemberRoleEnum string
const (
	FarmMemberOwner    FarmMemberRoleEnum = "owner"
	FarmMemberManager  FarmMemberRoleEnum = "manager"
	FarmMemberOperator FarmMemberRoleEnum = "operator"
	FarmMemberViewer   FarmMemberRoleEnum = "viewer"
)

// ── Device & Operational status ───────────────────────────────────────────────

type DeviceStatusEnum string
const (
	DeviceStatusOnline      DeviceStatusEnum = "online"
	DeviceStatusOffline     DeviceStatusEnum = "offline"
	DeviceStatusUnknown     DeviceStatusEnum = "unknown"
	DeviceStatusMaintenance DeviceStatusEnum = "maintenance"
	DeviceStatusError       DeviceStatusEnum = "error"
)

type OperationalStatusEnum string
const (
	OperationalStatusActive      OperationalStatusEnum = "active"
	OperationalStatusInactive    OperationalStatusEnum = "inactive"
	OperationalStatusMaintenance OperationalStatusEnum = "maintenance"
	OperationalStatusArchived    OperationalStatusEnum = "archived"
)

// ── Tasks ─────────────────────────────────────────────────────────────────────

type TaskStatusEnum string
const (
	TaskStatusPending    TaskStatusEnum = "pending"
	TaskStatusInProgress TaskStatusEnum = "in_progress"
	TaskStatusCompleted  TaskStatusEnum = "completed"
	TaskStatusCancelled  TaskStatusEnum = "cancelled"
	TaskStatusOverdue    TaskStatusEnum = "overdue"
)

// ── Notifications ─────────────────────────────────────────────────────────────

type NotificationPriorityEnum string
const (
	NotificationPriorityLow      NotificationPriorityEnum = "low"
	NotificationPriorityMedium   NotificationPriorityEnum = "medium"
	NotificationPriorityHigh     NotificationPriorityEnum = "high"
	NotificationPriorityCritical NotificationPriorityEnum = "critical"
)

type NotificationStatusEnum string
const (
	NotificationStatusPending        NotificationStatusEnum = "pending"
	NotificationStatusQueued         NotificationStatusEnum = "queued"
	NotificationStatusSent           NotificationStatusEnum = "sent"
	NotificationStatusDelivered      NotificationStatusEnum = "delivered"
	NotificationStatusFailedToSend   NotificationStatusEnum = "failed_to_send"
	NotificationStatusReadByUser     NotificationStatusEnum = "read_by_user"
	NotificationStatusAcknowledged   NotificationStatusEnum = "acknowledged_by_user"
	NotificationStatusArchivedByUser NotificationStatusEnum = "archived_by_user"
	NotificationStatusSystemCleared  NotificationStatusEnum = "system_cleared"
)

// ── Logging ───────────────────────────────────────────────────────────────────

type LogLevelEnum string
const (
	LogLevelDebug    LogLevelEnum = "debug"
	LogLevelInfo     LogLevelEnum = "info"
	LogLevelWarning  LogLevelEnum = "warning"
	LogLevelError    LogLevelEnum = "error"
	LogLevelCritical LogLevelEnum = "critical"
)

// ── Automation ────────────────────────────────────────────────────────────────

type AutomationTriggerSourceEnum string
const (
	AutomationTriggerSchedule  AutomationTriggerSourceEnum = "schedule"
	AutomationTriggerSensor    AutomationTriggerSourceEnum = "sensor_threshold"
	AutomationTriggerManual    AutomationTriggerSourceEnum = "manual"
	AutomationTriggerAPI       AutomationTriggerSourceEnum = "api"
	AutomationTriggerCondition AutomationTriggerSourceEnum = "condition_based"
)

type ExecutableActionTypeEnum string
const (
	ExecutableActionActuator     ExecutableActionTypeEnum = "actuator_command"
	ExecutableActionNotification ExecutableActionTypeEnum = "send_notification"
	ExecutableActionWebhook      ExecutableActionTypeEnum = "webhook"
	ExecutableActionTask         ExecutableActionTypeEnum = "create_task"
)

// ── Weather ───────────────────────────────────────────────────────────────────

type WeatherDataSourceEnum string
const (
	WeatherDataSourceSensor   WeatherDataSourceEnum = "on_site_sensor"
	WeatherDataSourceAPI      WeatherDataSourceEnum = "weather_api"
	WeatherDataSourceManual   WeatherDataSourceEnum = "manual_entry"
	WeatherDataSourceForecast WeatherDataSourceEnum = "forecast"
)

// ── Cost ──────────────────────────────────────────────────────────────────────

type CostCategoryEnum string
const (
	CostCategoryLabour      CostCategoryEnum = "labour"
	CostCategoryInputs      CostCategoryEnum = "inputs"
	CostCategoryEquipment   CostCategoryEnum = "equipment"
	CostCategoryInfrastructure CostCategoryEnum = "infrastructure"
	CostCategoryOther       CostCategoryEnum = "other"
)

// ── Farm scale ────────────────────────────────────────────────────────────────

type FarmScaleTierEnum string
const (
	FarmScaleMicro      FarmScaleTierEnum = "micro"
	FarmScaleSmall      FarmScaleTierEnum = "small"
	FarmScaleMedium     FarmScaleTierEnum = "medium"
	FarmScaleLarge      FarmScaleTierEnum = "large"
	FarmScaleEnterprise FarmScaleTierEnum = "enterprise"
)

// ── Validation ────────────────────────────────────────────────────────────────

type ValidationRuleTypeEnum string
const (
	ValidationRuleRange   ValidationRuleTypeEnum = "range"
	ValidationRuleRegex   ValidationRuleTypeEnum = "regex"
	ValidationRuleEnum    ValidationRuleTypeEnum = "enum"
	ValidationRuleCustom  ValidationRuleTypeEnum = "custom"
)

type ValidationSeverityEnum string
const (
	ValidationSeverityError   ValidationSeverityEnum = "error"
	ValidationSeverityWarning ValidationSeverityEnum = "warning"
	ValidationSeverityInfo    ValidationSeverityEnum = "info"
)

// ── User activity ─────────────────────────────────────────────────────────────

type UserActionTypeEnum string
const (
	UserActionCreate UserActionTypeEnum = "create"
	UserActionUpdate UserActionTypeEnum = "update"
	UserActionDelete UserActionTypeEnum = "delete"
	UserActionView   UserActionTypeEnum = "view"
	UserActionExport UserActionTypeEnum = "export"
	UserActionLogin  UserActionTypeEnum = "login"
	UserActionLogout UserActionTypeEnum = "logout"
)
