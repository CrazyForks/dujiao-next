package giftcard

// Service 礼品卡管理用例（不含兑换写路径）。
type Service struct {
	repo     Repository
	users    UserDirectory
	currency CurrencyProvider
}

// Options 组装管理用例依赖。
type Options struct {
	Repo     Repository
	Users    UserDirectory
	Currency CurrencyProvider
}

func NewService(opts Options) *Service {
	if opts.Repo == nil {
		panic("giftcard service: repo is nil")
	}
	return &Service{
		repo:     opts.Repo,
		users:    opts.Users,
		currency: opts.Currency,
	}
}
