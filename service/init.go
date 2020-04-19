package service

import "github.com/pot-code/go-di-pattern/container"

func Injection(dic *container.DIContainer) {
	dic.Register(new(LoginService))
	dic.Register(new(JWTService))
}

func InitPackage(dic *container.DIContainer) {
}
