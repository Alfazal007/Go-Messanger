package controllers

import (
	"database/sql"
	"encoding/json"
	helper "messager/helpers"
	"messager/internal/database"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (apiCfg *ApiConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	validate := validator.New()
	type parameter struct {
		Password string `json:"password" validate:"required,min=6"`
		Username string `json:"username" validate:"required,min=4"`
	}
	var params parameter
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		helper.RespondWithError(w, 400, "Error parsing the request body")
		return
	}
	err = validate.Struct(params)
	if err != nil {
		var errMsg string
		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Username" {
				errMsg += "Invalid username or email  "
			} else if err.Field() == "Password" {
				errMsg += "Password should be greater than or equal to 6 characters and less than or equal to 15  "
			}
		}
		helper.RespondWithError(w, 400, strings.Trim(errMsg, " "))
		return
	}

	user, err := apiCfg.DB.GetUserForLogin(r.Context(), params.Username)
	if err != nil {
		helper.RespondWithError(w, 400, "User not found in the database.")
		return
	}
	isPasswordCorrect := helper.CheckPasswordHash(params.Password, user.Password)
	if !isPasswordCorrect {
		helper.RespondWithError(w, 400, "Incorrect password")
		return
	}
	accessToken, err := GenerateJWT(user)
	if err != nil {
		helper.RespondWithError(w, 400, "Error generating access token")
		return
	}
	refreshToken, err := GenerateRefreshToken(user)
	if err != nil {
		helper.RespondWithError(w, 400, "Error generating refresh token")
		return
	}
	_, err = apiCfg.DB.UpdateRefreshToken(r.Context(), database.UpdateRefreshTokenParams{
		RefreshToken: sql.NullString{String: refreshToken, Valid: true},
		ID:           user.ID,
	})
	if err != nil {
		helper.RespondWithError(w, 400, "Error updating the database")
		return
	}

	cookie1 := http.Cookie{
		Name:     "access-token",
		Value:    accessToken,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie1)

	cookie2 := http.Cookie{
		Name:     "refresh-token",
		Value:    refreshToken,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie2)
	type AccessToken struct {
		AccessToken  string `json:"access-token"`
		RefreshToken string `json:"refresh-token"`
	}
	helper.RespondWithJSON(w, 200, AccessToken{AccessToken: accessToken, RefreshToken: refreshToken})
}
