package contract

// CheckInput 是订单风控检查所需的调用上下文。
type CheckInput struct {
	UserID      uint
	GuestEmail  string
	ClientIP    string
	IsGuest     bool
	SkipIPCheck bool
}
