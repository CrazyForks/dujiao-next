package provider

import (
	"context"

	adproxyapp "github.com/dujiao-next/internal/modules/adproxy/application"
	adproxygateway "github.com/dujiao-next/internal/modules/adproxy/infrastructure/adgateway"
	categoryapp "github.com/dujiao-next/internal/modules/catalog/category/application"

	giftcardintegration "github.com/dujiao-next/internal/integration/giftcard"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	cardsecretapp "github.com/dujiao-next/internal/modules/cardsecret/application"
	cartapp "github.com/dujiao-next/internal/modules/cart/application"
	contentapp "github.com/dujiao-next/internal/modules/content/application"
	"github.com/dujiao-next/internal/modules/content/infrastructure/gormstore"
	couponapp "github.com/dujiao-next/internal/modules/coupon/application"
	giftcardapp "github.com/dujiao-next/internal/modules/giftcard/application"
	giftcardsettingscurrency "github.com/dujiao-next/internal/modules/giftcard/infrastructure/settingscurrency"
	memberlevelapp "github.com/dujiao-next/internal/modules/memberlevel/application"
	orderriskapp "github.com/dujiao-next/internal/modules/orderrisk/application"
	orderrisklimiter "github.com/dujiao-next/internal/modules/orderrisk/infrastructure/redislimiter"
	promotionapp "github.com/dujiao-next/internal/modules/promotion/application"
	sitemapapp "github.com/dujiao-next/internal/modules/sitemap/application"
	sitemapcontract "github.com/dujiao-next/internal/modules/sitemap/contract"
	sitemapcache "github.com/dujiao-next/internal/modules/sitemap/infrastructure/cacheadapter"
	sitemapcatalog "github.com/dujiao-next/internal/modules/sitemap/infrastructure/catalogreader"
	walletapp "github.com/dujiao-next/internal/modules/wallet/application"
	"github.com/dujiao-next/internal/service"
)

// initApplicationServices 装配内容、购物车、订单、履约和营销用例。
func (c *Container) initApplicationServices() {
	c.AdProxyService = adproxyapp.NewService(adproxygateway.New())
	postStore := gormstore.NewPostStore(models.DB)
	postCategoryStore := gormstore.NewPostCategoryStore(models.DB)
	c.ContentPostService = contentapp.NewPostService(
		postStore,
		postStore,
		postCategoryStore,
		contentapp.SystemClock{},
	)
	c.ContentPostCategoryService = contentapp.NewPostCategoryService(postCategoryStore)
	c.CategoryService = categoryapp.NewService(c.CategoryRepo)
	sitemapService, err := sitemapapp.NewService(
		sitemapcatalog.New(c.ProductRepo, c.CategoryRepo),
		sitemapcontract.PublishedPostReaderFunc(func(ctx context.Context, limit int) ([]sitemapcontract.PublishedPost, error) {
			posts, _, listErr := c.ContentPostService.ListPublic(ctx, contentapp.PublicPostQuery{
				Page:     1,
				PageSize: limit,
			})
			if listErr != nil {
				return nil, listErr
			}
			result := make([]sitemapcontract.PublishedPost, 0, len(posts))
			for _, post := range posts {
				result = append(result, sitemapcontract.PublishedPost{
					Slug:        post.Slug,
					CreatedAt:   post.CreatedAt,
					PublishedAt: post.PublishedAt,
				})
			}
			return result, nil
		}),
		sitemapcache.New(),
	)
	if err != nil {
		logger.Errorw("provider_init_sitemap_failed", "error", err)
		panic(err)
	}
	c.SitemapService = sitemapService
	c.CartService = cartapp.NewService(c.CartRepo, c.ProductRepo, c.ProductSKURepo, c.PromotionRepo, c.SettingService)
	c.WalletService = walletapp.NewService(walletapp.Options{
		Repository: c.WalletRepo, Transactions: c.WalletRepo,
	})
	c.OrderRefundService = service.NewOrderRefundService(
		c.OrderRepo,
		c.UserStore,
		c.OrderRefundRecordRepo,
		c.AffiliateRefundHandler,
		c.SettingService,
		c.WalletService,
	)
	c.MemberLevelService = memberlevelapp.NewService(c.MemberLevelRepo, c.MemberLevelPriceRepo, c.MemberLevelUserRepo)
	c.OrderRiskControlService = orderriskapp.NewService(orderriskapp.Options{
		Settings:    c.SettingService,
		Orders:      c.OrderRepo,
		RateLimiter: orderrisklimiter.New(),
	})
	c.OrderService = service.NewOrderService(service.OrderServiceOptions{
		OrderRepo:               c.OrderRepo,
		OrderRefundRecordRepo:   c.OrderRefundRecordRepo,
		PaymentRepo:             c.PaymentRepo,
		UserStore:               c.UserStore,
		ProductRepo:             c.ProductRepo,
		ProductSKURepo:          c.ProductSKURepo,
		CardSecretRepo:          c.CardSecretRepo,
		ResellerStore:           c.ResellerStore,
		CouponRepo:              c.CouponRepo,
		CouponUsageRepo:         c.CouponUsageRepo,
		PromotionRepo:           c.PromotionRepo,
		QueueClient:             c.QueueClient,
		SettingService:          c.SettingService,
		DefaultEmailConfig:      c.Config.Email,
		WalletService:           c.WalletService,
		AffiliateService:        c.AffiliateService,
		MemberLevelService:      c.MemberLevelService,
		ResellerPricingResolver: c.ResellerPricingResolver,
		ResellerAccounting:      c.ResellerAccountingTransactions,
		RiskControlService:      c.OrderRiskControlService,
		ExpireMinutes:           c.Config.Order.PaymentExpireMinutes,
	})
	c.FulfillmentService = service.NewFulfillmentService(
		c.OrderRepo, c.FulfillmentRepo, c.CardSecretRepo, c.QueueClient,
		c.SettingService, c.Config.Email,
		c.ExternalIdentityStore,
	)
	c.CardSecretService = cardsecretapp.NewService(cardsecretapp.ServiceOptions{
		Secrets:      c.CardSecretRepo,
		Batches:      c.CardSecretBatchRepo,
		Transactions: c.CardSecretRepo,
		Products:     c.ProductRepo,
		ProductSKUs:  c.ProductSKURepo,
	})
	c.GiftCardService = giftcardapp.NewService(giftcardapp.Options{
		Repo:     c.GiftCardRepo,
		Users:    c.UserStore,
		Currency: giftcardsettingscurrency.New(c.SettingService),
		Redeemer: giftcardintegration.New(c.GiftCardRepo, c.WalletService),
	})
	c.CouponAdminService = couponapp.NewAdminService(c.CouponRepo)
	c.PromotionAdminService = promotionapp.NewAdminService(c.PromotionRepo)
	c.ContentBannerService = contentapp.NewBannerService(
		gormstore.NewBannerStore(models.DB),
		contentapp.SystemClock{},
	)
}
