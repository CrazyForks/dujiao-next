package uploadwiring

import (
	"github.com/dujiao-next/internal/provider"
	uploadtransport "github.com/dujiao-next/internal/transport/http/upload"
)

func NewAdminHandler(c *provider.Container) *uploadtransport.AdminHandler {
	return uploadtransport.NewAdminHandler(
		c.UploadService,
		c.ContentMediaService,
	)
}
