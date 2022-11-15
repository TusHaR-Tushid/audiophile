package middleware

import (
	"context"
	firebase "firebase.google.com/go"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"net/http"
	"os"
	"strings"
)

func OAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			logrus.Printf("OAuth:cannot create firebase application object:%v", err)
			return
		}
		auth, err := app.Auth(context.Background())
		if err != nil {
			logrus.Printf("OAuth:cannot create auth client:%v", err)
		}

		header := r.Header.Get(echo.HeaderAuthorization)
		idToken := strings.TrimSpace(strings.Replace(header, "Bearer", "", 1))
		_, err = auth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("cannot verify token:%v", err)
			return
		}
		ctx := context.WithValue(r.Context(), "id", idToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
