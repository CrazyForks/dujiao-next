package provider

import (
	"context"

	categoryapp "github.com/dujiao-next/internal/modules/catalog/category/application"

	giftcardintegration "github.com/dujiao-next/internal/integration/giftcard"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/cardsecret"
	"github.com/dujiao-next/internal/modules/cart"
	contentapp "github.com/dujiao-next/internal/modules/content/application"
	"github.com/dujiao-next/internal/modules/content/infrastructure/gormstore"
	couponapp "github.com/dujiao-next/internal/modules/coupon/application"
	giftcardapp "github.com/dujiao-next/internal/modules/giftcard/application"
	giftcardsettingscurrency "github.com/dujiao-next/internal/modules/giftcard/infrastructure/settingscurrency"
	memberlevelapp "github.com/dujiao-next/internal/modules/memberlevel/application"
	"github.com/dujiao-next/internal/modules/orderrisk"
	promotionapp "github.com/dujiao-next/internal/modules/promotion/application"
	"github.com/dujiao-next/internal/modules/sitemap"
	"github.com/dujiao-next/internal/service"
)

// initApplicationServices 装配内容、购物车、订单、履约和营销用例。
func (c *Container) initApplicationServices() {
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
	sitemapService, err := sitemap.NewService(
		c.ProductRepo,
		c.CategoryRepo,
		sitemap.PublishedPostReaderFunc(func(ctx context.Context, limit int) ([]sitemap.SitemapPost, error) {
			posts, _, listErr := c.ContentPostService.ListPublic(ctx, contentapp.PublicPostQuery{
				Page:     1,
				PageSize: limit,
			})
			if listErr != nil {
				return nil, listErr
			}
			result := make([]sitemap.SitemapPost, 0, len(posts))
			for _, post := range posts {
				result = append(result, sitemap.SitemapPost{
					Slug:        post.Slug,
					CreatedAt:   post.CreatedAt,
					PublishedAt: post.PublishedAt,
				})
			}
			return result, nil
		}),
	)
	if err != nil {
		logger.Errorw("provider_init_sitemap_failed", "error", err)
		panic(err)
	}
	c.SitemapService = sitemapService
	c.CartService = cart.NewService(c.CartRepo, c.ProductRepo, c.ProductSKURepo, c.PromotionRepo, c.SettingService)
	c.WalletService = service.NewWalletService(c.WalletRepo, c.OrderRepo, c.OrderRefundRecordRepo, c.UserStore, c.AffiliateRefundHandler, c.SettingService)
	c.OrderRefundService = service.NewOrderRefundService(c.OrderRepo, c.UserStore, c.OrderRefundRecordRepo, c.AffiliateRefundHandler, c.SettingService)
	c.MemberLevelService = memberlevelapp.NewService(c.MemberLevelRepo, c.MemberLevelPriceRepo, c.MemberLevelUserRepo)
	c.OrderRiskControlService = orderrisk.NewService(c.SettingService, c.OrderRepo)
	c.OrderService = service.NewOrderService(service.OrderServiceOptions{
		OrderRepo:                 c.OrderRepo,
		OrderRefundRecordRepo:     c.OrderRefundRecordRepo,
		PaymentRepo:               c.PaymentRepo,
		UserStore:                 c.UserStore,
		ProductRepo:               c.ProductRepo,
		ProductSKURepo:            c.ProductSKURepo,
		CardSecretRepo:            c.CardSecretRepo,
		ResellerRepo:              c.ResellerRepo,
		CouponRepo:                c.CouponRepo,
		CouponUsageRepo:           c.CouponUsageRepo,
		PromotionRepo:             c.PromotionRepo,
		QueueClient:               c.QueueClient,
		SettingService:            c.SettingService,
		DefaultEmailConfig:        c.Config.Email,
		WalletService:             c.WalletService,
		AffiliateService:          c.AffiliateService,
		MemberLevelService:        c.MemberLevelService,
		ResellerPricingResolver:   c.ResellerPricingResolver,
		ResellerAccountingService: c.ResellerAccountingService,
		RiskControlService:        c.OrderRiskControlService,
		ExpireMinutes:             c.Config.Order.PaymentExpireMinutes,
	})
	c.FulfillmentService = service.NewFulfillmentService(
		c.OrderRepo, c.FulfillmentRepo, c.CardSecretRepo, c.QueueClient,
		c.SettingService, c.Config.Email,
		c.ExternalIdentityStore,
	)
	c.CardSecretService = cardsecret.NewService(cardsecret.ServiceOptions{
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
