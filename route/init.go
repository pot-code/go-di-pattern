package route

import (
	"net/http"

	"github.com/pot-code/go-di-pattern/container"
)

func Injection(dic *container.DIContainer) {
	dic.Register(new(LoginController))
	dic.Register(new(JWTMiddleware))
}

func InitPackage(dic *container.DIContainer) {
	lc := dic.Get("LoginController").(*LoginController)
	jm := dic.Get("JWTMiddleware").(*JWTMiddleware)

	homeMiddleware := Compose(ErrorHandlingMiddleware, LoggingMiddleware, jm.ValidateMiddleware, jm.RefreshTokenMiddleware)
	loginMiddleware := Compose(ErrorHandlingMiddleware, LoggingMiddleware)
	logoutMiddleware := Compose(ErrorHandlingMiddleware, jm.ValidateMiddleware)

	http.HandleFunc("/login", loginMiddleware(lc.handleLogin))
	http.HandleFunc("/logout", logoutMiddleware(lc.handleLogout))
	http.HandleFunc("/home", homeMiddleware(lc.handleHome))
}
