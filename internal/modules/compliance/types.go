package compliance

// AcknowledgeRequest 确认请求。
type AcknowledgeRequest struct {
	Segment1  string
	Segment2  string
	Segment3  string
	AdminID   uint
	Username  string
	ClientIP  string
	UserAgent string
}

// Status 合规声明确认状态。
type Status struct {
	Acknowledged           bool   `json:"acknowledged"`
	AcknowledgedAt         string `json:"acknowledged_at,omitempty"`
	AcknowledgedByAdminID  uint   `json:"acknowledged_by_admin_id,omitempty"`
	AcknowledgedByUsername string `json:"acknowledged_by_username,omitempty"`
	Version                string `json:"version,omitempty"`
}
