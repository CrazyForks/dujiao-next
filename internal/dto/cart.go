package dto

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/jsonslice"
)

// CartProductResp 购物车商品摘要
type CartProductResp struct {
	Slug                string             `json:"slug"`
	Title               jsonmap.JSON       `json:"title"`
	PriceAmount         models.Money       `json:"price_amount"`
	Images              jsonslice.Strings `json:"images"`
	Tags                jsonslice.Strings `json:"tags"`
	PurchaseType        string             `json:"purchase_type"`
	MinPurchaseQuantity int                `json:"min_purchase_quantity"`
	MaxPurchaseQuantity int                `json:"max_purchase_quantity"`
	FulfillmentType     string             `json:"fulfillment_type"`
	IsActive            bool               `json:"is_active"`
}

// CartItemResp 购物车项响应
type CartItemResp struct {
	ProductID       uint            `json:"product_id"`
	SKUID           uint            `json:"sku_id"`
	Quantity        int             `json:"quantity"`
	FulfillmentType string          `json:"fulfillment_type"`
	UnitPrice       models.Money    `json:"unit_price"`
	OriginalPrice   models.Money    `json:"original_price"`
	Currency        string          `json:"currency"`
	Product         CartProductResp `json:"product"`
}
