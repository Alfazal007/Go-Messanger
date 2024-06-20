package controllers

import (
	"database/sql"
	"fmt"
	helper "messager/helpers"
	"messager/internal/database"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func (apiCfg *ApiConfig) UpdateRefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh-token")
	var refreshToken string
	if err != nil {
		if err == http.ErrNoCookie {
			authorization := r.Header.Get("Authorization")
			if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
				http.Error(w, "Authorization header missing or improperly formatted", http.StatusUnauthorized)
				helper.RespondWithError(w, 400, "No headers provided")
				return
			}
			refreshToken = strings.TrimPrefix(authorization, "Bearer ")
		} else {
			helper.RespondWithError(w, 400, "Error reading cookie, try logging in again")
			return
		}
	} else {
		refreshToken = cookie.Value
	}
	// Verify the JWT token
	jwtSecret := os.Getenv("SECRET_KEY_REFRESH")
	if jwtSecret == "" {
		helper.RespondWithError(w, 400, "Server error")
		return
	}
	if refreshToken == "" {
		helper.RespondWithError(w, 400, "Provide cookie")
		return
	}
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		helper.RespondWithError(w, 401, fmt.Sprintf("Invalid token here %v", err))
		return
	}
	if !token.Valid {
		helper.RespondWithError(w, 401, fmt.Sprintf("Invalid token %v", err))
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		helper.RespondWithError(w, 400, "Invalid claims login again")
		return
	}

	email := claims["email"].(string)
	userId := claims["user_id"].(string)
	userFromDB, err := apiCfg.DB.GetUserByEmail(r.Context(), email)
	if err != nil {
		helper.RespondWithError(w, 400, "No user in the database")
		return
	}
	if userFromDB.RefreshToken.String != refreshToken {
		helper.RespondWithError(w, 400, "Invalid refresh token")
		return
	}
	id, err := uuid.Parse(userId)
	if err != nil {
		helper.RespondWithError(w, 400, "Error parsing user id")
		return
	}
	if id != userFromDB.ID {
		helper.RespondWithError(w, 400, "Invalid user")
		return
	}
	accessToken, err := GenerateJWT(userFromDB)
	if err != nil {
		helper.RespondWithError(w, 400, "Error generating access token")
		return
	}
	refreshToken, err = GenerateRefreshToken(userFromDB)
	if err != nil {
		helper.RespondWithError(w, 400, "Error generating refresh token")
		return
	}
	_, err = apiCfg.DB.UpdateRefreshToken(r.Context(), database.UpdateRefreshTokenParams{
		RefreshToken: sql.NullString{String: refreshToken, Valid: true},
		ID:           userFromDB.ID,
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
