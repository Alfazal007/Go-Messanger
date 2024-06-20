package helper

import (
	"messager/internal/database"
	"time"

	"github.com/google/uuid"
)

type CustomUser struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	ID        uuid.UUID `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ConvertToUser(user database.User) CustomUser {
	return CustomUser{
		Username:  user.Username,
		Email:     user.Email,
		ID:        user.ID,
		UpdatedAt: user.UpdatedAt,
	}
}
