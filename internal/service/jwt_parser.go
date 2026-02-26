package service

import "github.com/golang-jwt/jwt/v5"

// newHS256JWTParser 统一 HS256 JWT 解析器，避免不同调用点出现方法白名单偏差。
func newHS256JWTParser() *jwt.Parser {
	return jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
}
