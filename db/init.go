package db

import (
	"github.com/pot-code/go-di-pattern/container"
)

func Injection(dic *container.DIContainer) {
	dic.Register(new(RedisDB))
}

func InitPackage(dic *container.DIContainer) {
}
