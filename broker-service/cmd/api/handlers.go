package main

import (
	"net/http"

	"github.com/crux25/go-compmgt/helpers"
)

func (app *Config) broker(w http.ResponseWriter, r *http.Request) {
	payload := helpers.JSONResponse{
		Error:   false,
		Message: "The broker is reachable",
	}

	var helper = new(helpers.JSONHelper)
	_ = helper.WriteJSON(w, http.StatusOK, payload)
}
