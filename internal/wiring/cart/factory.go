package cartwiring

import (
	"github.com/dujiao-next/internal/provider"
	carttransport "github.com/dujiao-next/internal/transport/http/cart"
)

func NewUserHandler(c *provider.Container) *carttransport.UserHandler {
	return carttransport.NewUserHandler(c.CartService)
}
