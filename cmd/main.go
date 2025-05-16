package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Se623/calc-full-app/internal/agent"
	"github.com/Se623/calc-full-app/internal/database"
	"github.com/Se623/calc-full-app/internal/lib"
	"github.com/Se623/calc-full-app/internal/manager"
	"github.com/Se623/calc-full-app/internal/orchestrator"
	"github.com/golang-jwt/jwt/v5"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { return lib.SecretKey, nil })
		if err != nil || !token.Valid {
			http.Error(w, "Token invalid", http.StatusUnauthorized)
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		id := int(claims["id"].(float64))
		fmt.Println(id)
		name := claims["name"].(string)
		pass := claims["pass"].(string)
		_, err = database.DBM.CheckUser(name, pass)
		if err != nil {
			http.Error(w, "Token invalid", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "id", id)
		next(w, r.WithContext(ctx))
	}
}

func main() {
	lib.InitLogger()

	lib.Sugar.Info("Initilized main program")

	for i := 0; i < lib.COMPUTING_AGENTS; i++ {
		go agent.Agent(i + 1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", manager.Registator)
	mux.HandleFunc("/api/v1/login", manager.Loginer)
	mux.HandleFunc("/api/v1/calculate", authMiddleware(orchestrator.Spliter))
	mux.HandleFunc("localhost/api/v1/expressions", authMiddleware(orchestrator.Displayer))
	go orchestrator.Distributor()
	lib.Sugar.Infof("Initilized orchestrator")

	err := database.Init()
	if err != nil {
		lib.Sugar.Errorf(err.Error())
		return
	}
	lib.Sugar.Infof("Initilized database")

	if err = http.ListenAndServe(":8080", mux); err != nil { // Запуск сервера
		lib.Sugar.Errorf(err.Error())
		return
	}
	lib.Sugar.Infof("Initilized server on port 8080")
}
