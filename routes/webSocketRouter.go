package routes

import (
	"messager/controllers"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func WebsocketRouter(apiCfg *controllers.ApiConfig) *chi.Mux {
	r := chi.NewRouter()
	m := controllers.NewManager()
	r.Post("/ws", controllers.VerifyJWT(apiCfg, http.HandlerFunc(m.ServeWebSocket)).ServeHTTP)
	return r
}
