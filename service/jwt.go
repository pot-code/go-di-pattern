package service

import (
	"github.com/dgrijalva/jwt-go"
)

type AppTokenClaims struct {
	Name string `json:"name"` // username
	jwt.StandardClaims
}

type IJWTService interface {
	Sign(claims jwt.Claims) (string, error)
	Validate(tokenStr string, claims jwt.Claims) (*jwt.Token, error)
}

type JWTService struct {
	secret []byte
	method jwt.SigningMethod
}

func (manager JWTService) Constructor() *JWTService {
	return &JWTService{[]byte("jwtdemo"), jwt.SigningMethodHS256}
}

func (manager *JWTService) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(manager.method, claims)
	tokenStr, err := token.SignedString(manager.secret)
	return tokenStr, err
}

func (manager *JWTService) Validate(tokenStr string, claims jwt.Claims) (*jwt.Token, error) {
	parseToken, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return manager.secret, nil
	})
	return parseToken, err
}
