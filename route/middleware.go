package route

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pot-code/go-di-pattern/service"

	"github.com/dgrijalva/jwt-go"
)

const SessionTimeout = 30 * time.Minute
const RefreshTokenThreshold = 5 * time.Minute

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc
type JWTToken string

type JWTMiddleware struct {
	*service.LoginService `dep:"LoginService"`
	*service.JWTService   `dep:"JWTService"`
}

func (lm JWTMiddleware) Constructor() *JWTMiddleware {
	return &JWTMiddleware{lm.LoginService, lm.JWTService}
}

// ValidateMiddleware validate JWT
func (lm *JWTMiddleware) ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if tokenStr, err := lm.GetToken(req); err == nil {
			token, err := lm.Validate(tokenStr, &service.AppTokenClaims{})

			if err == nil {
				if !lm.IsInvalidToken(tokenStr) {
					req = req.WithContext(context.WithValue(req.Context(), JWTToken("jwt-token"), token))
					next(res, req)
					return
				}
				log.Printf("[%s] %s %s invalidated token %s", req.Method, req.RequestURI, req.RemoteAddr, tokenStr)
			} else {
				log.Printf("[%s] %s %s %v", req.Method, req.RequestURI, req.RemoteAddr, err)
			}
		} else {
			log.Printf("[%s] %s %s %v", req.Method, req.RequestURI, req.RemoteAddr, err)
		}
		res.WriteHeader(http.StatusUnauthorized)
	}
}

func (lm *JWTMiddleware) RefreshTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token := req.Context().Value(JWTToken("jwt-token")).(*jwt.Token)
		claims := token.Claims.(*service.AppTokenClaims)
		exp := claims.ExpiresAt

		if time.Unix(exp, 0).Sub(time.Now()) < RefreshTokenThreshold {
			claims.ExpiresAt = time.Now().Add(SessionTimeout).Unix()
			newToken, _ := lm.Sign(claims)
			lm.SetToken(res, newToken, time.Now().Add(SessionTimeout))
		}
		next(res, req)
	}
}

// LoggingMiddleware log request
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Printf("[%s] %s", req.Method, req.RequestURI)
		next(res, req)
	}
}

// Compose compose a set of middlewares to a single one
func Compose(first MiddlewareFunc, rest ...MiddlewareFunc) MiddlewareFunc {
	acc := first
	n := len(rest)

	for i := 0; i < n; i++ {
		acc = func(idx int, prev MiddlewareFunc) MiddlewareFunc {
			return func(next http.HandlerFunc) http.HandlerFunc {
				return prev(rest[idx](next))
			}
		}(i, acc) // closure problem
	}
	return acc
}

// ErrorHandlingMiddleware handle panic
func ErrorHandlingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[%s] %s %s %v", req.Method, req.RequestURI, req.RemoteAddr, err)
				res.Header().Set("Content-Type", "application/json")
				body, _ := json.Marshal(ReturnMessage{
					Status:  false,
					Message: fmt.Sprintf("Error: %v", err),
				})
				res.WriteHeader(http.StatusInternalServerError)
				res.Write(body)
			}
		}()
		next(res, req)
	}
}
