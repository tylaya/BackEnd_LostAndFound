package routes

import (
	"backend-lostfound/handlers"
	"backend-lostfound/middleware"
	"net/http"
)

func RegisterRoutes() {
	// Rute Publik (Tanpa Token)
	http.HandleFunc("/api/register", handlers.Register)
	http.HandleFunc("/api/login", handlers.Login)

	// Rute VIP (Wajib Token JWT)
	http.HandleFunc("/api/profile", middleware.RequireAuth(handlers.Profile))
	http.HandleFunc("/api/change-password", middleware.RequireAuth(handlers.ChangePassword))
	http.HandleFunc("/api/status-barang", middleware.RequireAuth(handlers.UpdateStatusBarang))
}
