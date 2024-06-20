package controllers

import (
	"context"
	"fmt"
	"log"
	helper "messager/helpers"
	"messager/internal/database"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func GenerateJWT(user database.User) (string, error) {
	jwtSecret := os.Getenv("SECRET_KEY")
	if jwtSecret == "" {
		log.Fatal("DB url is not found in env variables")
	}
	secretKey := []byte(jwtSecret)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["authorized"] = true
	claims["user_id"] = user.ID
	claims["email"] = user.Email

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GenerateRefreshToken(user database.User) (string, error) {
	jwtSecret := os.Getenv("SECRET_KEY_REFRESH")
	if jwtSecret == "" {
		log.Fatal("DB url is not found in env variables")
	}
	secretKey := []byte(jwtSecret)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(100 * time.Hour).Unix()
	claims["authorized"] = true
	claims["user_id"] = user.ID
	claims["email"] = user.Email
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyJWT(apiCfg *ApiConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access-token")
		var jwtToken string
		if err != nil {
			if err == http.ErrNoCookie {
				authorization := r.Header.Get("Authorization")
				if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
					http.Error(w, "Authorization header missing or improperly formatted", http.StatusUnauthorized)
					helper.RespondWithError(w, 400, "No headers provided")
					return
				}

				jwtToken = strings.TrimPrefix(authorization, "Bearer ")
			} else {
				helper.RespondWithError(w, 400, "Error reading cookie, try logging in again")
				return
			}
		} else {
			jwtToken = cookie.Value
		}
		// Verify the JWT token
		jwtSecret := os.Getenv("SECRET_KEY")
		if jwtSecret == "" {
			helper.RespondWithError(w, 400, "Server error")
			return
		}
		if jwtToken == "" {
			helper.RespondWithError(w, 400, "Provide cookie")
			return
		}
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
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
		id := claims["user_id"].(string)

		idInUUID, err := uuid.Parse(id)
		if err != nil {
			helper.RespondWithError(w, 400, "Token has been tampered with login again")
			return
		}

		user, err := apiCfg.DB.GetUserById(r.Context(), idInUUID)
		if err != nil {
			helper.RespondWithError(w, 400, "Some manpulation done with the token or the user has been deleted probably")
			return
		}
		if email != user.Email {
			helper.RespondWithError(w, 400, "Some manipulations done with the token try again")
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
