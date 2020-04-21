package route

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pot-code/go-di-pattern/service"
)

type ReturnMessage struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type LoginController struct {
	*JWTMiddleware `dep:"JWTMiddleware"`
}

func (c LoginController) Constructor() *LoginController {
	instance := &LoginController{c.JWTMiddleware}

	homeMiddleware := Compose(ErrorHandlingMiddleware, LoggingMiddleware, c.ValidateMiddleware, c.RefreshTokenMiddleware)
	loginMiddleware := Compose(ErrorHandlingMiddleware, LoggingMiddleware)
	logoutMiddleware := Compose(ErrorHandlingMiddleware, c.ValidateMiddleware)

	http.HandleFunc("/login", loginMiddleware(c.handleLogin))
	http.HandleFunc("/logout", logoutMiddleware(c.handleLogout))
	http.HandleFunc("/home", homeMiddleware(c.handleHome))
	return instance
}

func (c *LoginController) handleHome(res http.ResponseWriter, req *http.Request) {
	header := res.Header()
	header.Set("Content-Type", "text/html")
	io.WriteString(res, `
  <html>
    <body>
      <h1>Welcome home</h1>
    </body>
  </html>`)
}

func (c *LoginController) handleLogout(res http.ResponseWriter, req *http.Request) {
	tokenStr, _ := c.GetToken(req)
	token := req.Context().Value(JWTToken("jwt-token")).(*jwt.Token)

	c.LoginService.InvalidateToken(token, tokenStr)
	http.SetCookie(res, &http.Cookie{
		Name:   "auth-token",
		Value:  "",
		Path:   "",
		MaxAge: -1,
	})
}

func (c *LoginController) handleLogin(res http.ResponseWriter, req *http.Request) {
	header := res.Header()

	if req.Method != http.MethodPost {
		header.Set("Content-Type", "application/json")
		body, _ := json.Marshal(ReturnMessage{
			Status:  false,
			Message: fmt.Sprintf("invalid HTTP method, expected: %s, actual: %s", http.MethodPost, req.Method),
		})
		res.Write(body)
		return
	}

	username := req.FormValue("username")
	if username == "" {
		header.Set("Content-Type", "application/json")
		body, _ := json.Marshal(ReturnMessage{
			Status:  false,
			Message: "username is empty",
		})
		res.Write(body)
		return
	}

	tokenStr, _ := c.Sign(service.AppTokenClaims{
		Name: "demo",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(SessionTimeout).Unix(),
		},
	})
	c.SetToken(res, tokenStr, time.Now().Add(SessionTimeout))
	res.WriteHeader(http.StatusOK)
}
