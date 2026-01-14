package main

import (
	"errors"
	"net/http"

	"github.com/mightyfzeus/rbac/cmd/helpers"
	"github.com/mightyfzeus/rbac/internal/dtos"
)

func (app *application) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload dtos.LoginPayload

	if err := app.DecodeAndValidate(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.User.LoginUser(r.Context(), payload.Email, payload.Password)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if user.Status == helpers.StatusPending {
		app.badRequestResponse(w, r, errors.New("user is pending"))
		return
	}

	token, err := GenerateJWT(user.ID, user.Email, user.Name, user.Role)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"data":  user,
	}, "login successful")

}
