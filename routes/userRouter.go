package routes

import (
	"messager/controllers"

	"github.com/go-chi/chi/v5"
)

func UserRouter(apiCfg *controllers.ApiConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", apiCfg.CreateUser)
	r.Post("/login", apiCfg.LoginUser)
	return r
}
