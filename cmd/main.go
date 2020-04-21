package main

import (
	"log"
	"net/http"

	"github.com/pot-code/go-di-pattern/db"
	"github.com/pot-code/go-di-pattern/service"

	"github.com/pot-code/go-di-pattern/container"

	"github.com/pot-code/go-di-pattern/route"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[jwt demo]")

	dic := container.NewDIContainer()
	dic.Register(new(route.LoginController))
	dic.Register(new(route.JWTMiddleware))
	dic.Register(new(db.RedisDB))
	dic.Register(new(service.LoginService))
	dic.Register(new(service.JWTService))

	dic.Get("LoginController") // trigger the initialization
	http.ListenAndServe(":8080", nil)
}
