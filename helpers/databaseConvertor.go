package helper

import (
	"messager/internal/database"
	"time"

	"github.com/google/uuid"
)

type CustomUser struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	ID        uuid.UUID `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ConvertToUser(user database.User) CustomUser {
	return CustomUser{
		Name:      user.Name,
		Email:     user.Email,
		ID:        user.ID,
		UpdatedAt: user.UpdatedAt,
	}
}
