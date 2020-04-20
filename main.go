package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pot-code/go-di-pattern/db"

	"github.com/pot-code/go-di-pattern/service"

	"github.com/pot-code/go-di-pattern/container"

	"github.com/pot-code/go-di-pattern/route"
)

type Application struct {
	dic      *container.DIContainer
	injected bool
}

func NewApplication() *Application {
	return &Application{
		dic:      container.NewDIContainer(),
		injected: false,
	}
}

func (app *Application) InjectDependency() {
	if app.injected {
		return
	}
	dic := app.dic
	route.Injection(dic)
	service.Injection(dic)
	db.Injection(dic)

	route.InitPackage(dic)
	service.InitPackage(dic)
	db.InitPackage(dic)
	app.injected = true
}

func (app *Application) Serve(host string, mux http.Handler) {
	if !app.injected {
		app.InjectDependency()
	}
	http.ListenAndServe(host, mux)
}

func (app *Application) GetComponent(name string) (interface{}, error) {
	if !app.injected {
		return nil, fmt.Errorf("Haven't injected yet")
	}
	return app.dic.Get(name), nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[jwt demo]")

	app := NewApplication()
	app.Serve(":8080", nil)
}
