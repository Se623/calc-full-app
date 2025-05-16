package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Se623/calc-full-app/internal/database"
	"github.com/Se623/calc-full-app/internal/lib"
	"github.com/golang-jwt/jwt/v5"
)

func Registator(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var resp lib.Cred
	err := decoder.Decode(&resp)

	if err != nil {
		http.Error(w, "Error: Invalid JSON", http.StatusInternalServerError)
		return
	}
	_, err = database.DBM.CheckUser(resp.Login, resp.Password)
	if err == nil {
		http.Error(w, "Error: User exists", http.StatusTeapot)
		return
	}
	id, err := database.DBM.AddUser(resp.Login, resp.Password)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `{"id": "%s"}`, strconv.Itoa(id))
}

func Loginer(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var resp lib.Cred
	err := decoder.Decode(&resp)

	if err != nil {
		http.Error(w, "Error: Invalid JSON", http.StatusInternalServerError)
		return
	}

	id, err := database.DBM.CheckUser(resp.Login, resp.Password)
	if err != nil {
		http.Error(w, "Error: Unknown user", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"id":   id,
		"name": resp.Login,
		"pass": resp.Password,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(lib.SecretKey)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, `{"token": "%s"}`, string(signed))
}
