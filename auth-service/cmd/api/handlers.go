package main

import (
	_ "crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/crux25/go-compmgt/helpers"
	"github.com/golang-jwt/jwt"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.jsonHelper.ReadJSON(w, r, &requestPayload)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// Generate JWT token
	jwtToke, err := app.generateJWT(user.Email)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	respData := struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Token     string `json:"token,omitempty"`
	}{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Token:     jwtToke,
	}

	payload := helpers.JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    respData,
	}

	app.jsonHelper.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) generateJWT(user string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = float64(time.Now().Add(60 * time.Minute).Unix()) // This token should expire after 1 hour.
	claims["user"] = user
	tokenString, err := token.SignedString([]byte(os.Getenv("SecretKey")))
	if err != nil {
		return "Signing Error", err
	}
	return tokenString, nil

}

func (app *Config) Validate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Token string `json:"token"`
	}

	err := app.jsonHelper.ReadJSON(w, r, &requestPayload)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(requestPayload.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SecretKey")), nil

	})

	if err != nil || !token.Valid {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	email := claims["user"].(string)
	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	respData := struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Token     string `json:"token,omitempty"`
	}{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Token:     requestPayload.Token,
	}

	payload := helpers.JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    respData,
	}

	app.jsonHelper.WriteJSON(w, http.StatusAccepted, payload)
}
