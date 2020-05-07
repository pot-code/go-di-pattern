package main

import (
	"log"
	"net/http"

	"github.com/pot-code/go-di-pattern/service"
	"github.com/pot-code/go-injection"

	"github.com/pot-code/go-di-pattern/route"
)

type App struct {
	LoginController *route.LoginController `dep:""`
	Host            string
}

func (app App) Constructor() *App {
	if err := http.ListenAndServe(app.Host, nil); err != nil {
		log.Fatal(err)
	}
	return &app
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[go-di-pattern]")

	app := &App{Host: ":8080"}
	dic := injection.NewDIContainer()
	dic.Register(new(route.LoginController))
	dic.Register(new(service.LoginService))
	dic.Register(new(service.JWTService))
	dic.Register(app)

	err := dic.Populate()
	if err != nil {
		log.Fatal(err)
	}
}
