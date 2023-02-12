package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/crux25/go-compmgt/helpers"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) broker(w http.ResponseWriter, r *http.Request) {
	payload := helpers.JSONResponse{
		Error:   false,
		Message: "The broker is reachable",
	}

	var helper = new(helpers.JSONHelper)
	_ = helper.WriteJSON(w, http.StatusOK, payload)
}

func (app *Config) handleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.jsonHelper.ReadJSON(w, r, &requestPayload)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	default:
		app.jsonHelper.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://auth-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.jsonHelper.ErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.jsonHelper.ErrorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.jsonHelper.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create a varabiel we'll read response.Body into
	var jsonFromService helpers.JSONResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.jsonHelper.ErrorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.jsonHelper.ErrorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload helpers.JSONResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.jsonHelper.WriteJSON(w, http.StatusAccepted, payload)
}
