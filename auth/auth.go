// Package auth is authentication for test.
// Don't use release code.
package auth

import (
	"net/http"
	"time"

	middleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"
)

// GetTokenHandler get token
var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// headerのセット
	token := jwt.New(jwt.SigningMethodHS256)

	// claimsのセット
	claims := token.Claims.(jwt.MapClaims)
	claims["admin"] = true
	claims["name"] = "turanukimaru"
	claims["iat"] = time.Now().Unix() // Unix() を入れ忘れると Error parsing token: Token used before issued
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// 電子署名
	tokenString, _ := token.SignedString([]byte("SECRET"))

	// JWTを返却
	w.Write([]byte(tokenString))
})

// JwtMiddleware check token
var JwtMiddleware = middleware.New(middleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte("SECRET"), nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})
