package service

import (
	"time"

	"github.com/pot-code/go-di-pattern/db"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
)

type LoginService struct {
	RedisClient *db.RedisDB `dep:""`
}

func (ls LoginService) Constructor() *LoginService {
	return &LoginService{ls.RedisClient}
}

func (ls *LoginService) InvalidateToken(token *jwt.Token, tokenStr string) {
	client := ls.RedisClient
	claims := token.Claims.(*AppTokenClaims)
	dur := time.Unix(claims.ExpiresAt, 0).Sub(time.Now())

	client.Set(tokenStr, true, dur)
}

func (ls *LoginService) IsInvalidToken(tokenStr string) bool {
	client := ls.RedisClient
	_, err := client.Get(tokenStr).Result()

	if err == redis.Nil {
		return false
	} else if err != nil {
		panic(err)
	}
	return true
}
