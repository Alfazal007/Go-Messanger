package controllers

import (
	"encoding/json"
	"fmt"
	helper "messager/helpers"
	"messager/internal/database"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func (apiCf *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	validate := validator.New()
	type parameter struct {
		Username string `json:"username" validate:"required,min=6,max=15"`
		Password string `json:"password" validate:"required,min=6"`
		Email    string `json:"email" validate:"required,email"`
	}
	var params parameter
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		helper.RespondWithError(w, 400, fmt.Sprintf("Error parsing JSON %v", err))
		return
	}

	err = validate.Struct(params)
	if err != nil {
		var errMsg string
		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Email" {
				errMsg += "Invalid email  "
			} else if err.Field() == "Username" {
				errMsg += "Username should be greater than or equal to 6 characters and less than or equal to 15  "
			} else if err.Field() == "Password" {
				errMsg += "Password should be greater than or equal to 6 characters and less than or equal to 15  "
			}
		}
		helper.RespondWithError(w, 400, strings.Trim(errMsg, " "))
		return
	}
	hashedPassword, err := helper.HashPassword(params.Password)
	if err != nil {
		helper.RespondWithError(w, 400, "Error hashing the password")
		return
	}
	userCount, err := apiCf.DB.GetUserBeforeCreate(r.Context(), database.GetUserBeforeCreateParams{
		Username: params.Username,
		Email:    params.Email,
	})
	if err != nil {
		helper.RespondWithError(w, 400, "Error talking to the database")
		return
	}
	if userCount != 0 {
		helper.RespondWithError(w, 400, "User with similar email or password already exists in the database.")
		return
	}
	user, err := apiCf.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		Username:  params.Username,
		Password:  hashedPassword,
		Email:     params.Email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		fmt.Println(err)
		helper.RespondWithError(w, 400, "Error creating the user in the database")
		return
	}
	helper.RespondWithJSON(w, 201, helper.ConvertToUser(user))
}
