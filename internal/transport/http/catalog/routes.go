package cataloghttp

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes 注册公开商品目录端点。
func RegisterPublicRoutes(public gin.IRoutes, handler *PublicHandler) {
	if public == nil || handler == nil {
		panic("catalog public routes: required dependency is nil")
	}
	public.GET("/products", handler.GetProducts)
	public.GET("/products/:slug", handler.GetProductBySlug)
	public.GET("/categories", handler.GetCategories)
}

// RegisterAdminCategoryRoutes 注册商品分类后台端点。
func RegisterAdminCategoryRoutes(admin gin.IRoutes, handler *AdminCategoryHandler) {
	if admin == nil || handler == nil {
		panic("catalog admin category routes: required dependency is nil")
	}
	admin.GET("/categories", handler.GetAdminCategories)
	admin.POST("/categories", handler.CreateCategory)
	admin.PUT("/categories/:id", handler.UpdateCategory)
	admin.PATCH("/categories/:id/active", handler.PatchCategoryActive)
	admin.DELETE("/categories/:id", handler.DeleteCategory)
}

// RegisterAdminProductRoutes 注册商品后台端点。
func RegisterAdminProductRoutes(admin gin.IRoutes, handler *AdminProductHandler) {
	if admin == nil || handler == nil {
		panic("catalog admin product routes: required dependency is nil")
	}
	admin.GET("/products", handler.GetAdminProducts)
	admin.GET("/products/:id", handler.GetAdminProduct)
	admin.POST("/products", handler.CreateProduct)
	admin.PUT("/products/:id", handler.UpdateProduct)
	admin.PATCH("/products/:id/wholesale-prices", handler.UpdateProductWholesalePrices)
	admin.PATCH("/products/:id", handler.QuickUpdateProduct)
	admin.DELETE("/products/:id", handler.DeleteProduct)
	admin.POST("/products/batch-status", handler.BatchUpdateProductStatus)
	admin.POST("/products/batch-category", handler.BatchUpdateProductCategory)
	admin.POST("/products/batch-delete", handler.BatchDeleteProducts)
}

// RegisterAdminProductMappingRoutes 注册商品映射后台端点。
func RegisterAdminProductMappingRoutes(admin gin.IRoutes, handler *AdminProductMappingHandler) {
	if admin == nil || handler == nil {
		panic("catalog admin product mapping routes: required dependency is nil")
	}
	admin.GET("/product-mappings", handler.GetProductMappings)
	admin.GET("/product-mappings/:id", handler.GetProductMapping)
	admin.POST("/product-mappings/import", handler.ImportUpstreamProduct)
	admin.POST("/product-mappings/batch-import", handler.BatchImportUpstreamProducts)
	admin.POST("/product-mappings/:id/sync", handler.SyncProductMapping)
	admin.PUT("/product-mappings/:id/status", handler.UpdateProductMappingStatus)
	admin.DELETE("/product-mappings/:id", handler.DeleteProductMapping)
	admin.POST("/product-mappings/batch-sync", handler.BatchSyncProductMappings)
	admin.POST("/product-mappings/batch-status", handler.BatchUpdateProductMappingStatus)
	admin.POST("/product-mappings/batch-delete", handler.BatchDeleteProductMappings)
	admin.GET("/upstream-products", handler.ListUpstreamProducts)
	admin.GET("/upstream-categories", handler.ListUpstreamCategories)
	admin.POST("/product-mappings/batch-import-by-category", handler.BatchImportByCategory)
}
