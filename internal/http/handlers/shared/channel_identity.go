package shared

import (
	externalidentitydomain "github.com/dujiao-next/internal/modules/identity/externalidentity/domain"
	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	"github.com/gin-gonic/gin"
)

// BuildChannelIdentityResponse 构造 Telegram 渠道身份响应载荷。
func BuildChannelIdentityResponse(bound, created bool, user *userdomain.User, identity *externalidentitydomain.Identity) gin.H {
	resp := gin.H{
		"bound": bound,
	}
	if identity != nil {
		resp["identity"] = gin.H{
			"provider":         identity.Provider,
			"provider_user_id": identity.ProviderUserID,
			"username":         identity.Username,
			"avatar_url":       identity.AvatarURL,
		}
	}
	if user != nil {
		resp["user"] = gin.H{
			"id":                      user.ID,
			"email":                   user.Email,
			"display_name":            user.DisplayName,
			"status":                  user.Status,
			"locale":                  user.Locale,
			"email_verified":          user.EmailVerifiedAt != nil,
			"password_setup_required": user.PasswordSetupRequired,
		}
	}
	if bound {
		resp["created"] = created
	}
	return resp
}
