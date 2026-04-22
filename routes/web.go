package routes

import (
	"net/http"

	"backend-lostfound/handlers"
	"backend-lostfound/middleware"
)

func RegisterRoutes() {
	// Rute Publik (Siapa saja boleh akses tanpa tiket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "🚀 Server API Lost & Found Berjalan dengan Baik!"}`))
	})
	http.HandleFunc("/api/register", handlers.Register)
	http.HandleFunc("/api/login", handlers.Login)

	// Rute Rahasia / VIP (Wajib pakai tiket JWT)
	// Perhatikan bagaimana middleware.RequireAuth membungkus handlers.Profile
	http.HandleFunc("/api/profile", middleware.RequireAuth(handlers.Profile))
}
