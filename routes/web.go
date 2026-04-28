// routes/web.go
package routes

import (
	"backend-lostfound/config"
	"backend-lostfound/handlers"
	"backend-lostfound/middleware"
	"net/http"
)

func RegisterRoutes() {
	// Rute Publik (Tanpa Token)
	http.HandleFunc("/api/register", handlers.Register)
	http.HandleFunc("/api/login", handlers.Login)

	// Rute Ambil Gambar 
	http.HandleFunc("/api/image", handlers.GetImageHandler(config.DB))

	// Rute VIP (Wajib Token JWT)
	http.HandleFunc("/api/profile", middleware.RequireAuth(handlers.Profile))
	http.HandleFunc("/api/change-password", middleware.RequireAuth(handlers.ChangePassword))
	http.HandleFunc("/api/status-barang", middleware.RequireAuth(handlers.UpdateStatusBarang))
	
	http.HandleFunc("/api/barang/hilang", middleware.RequireAuth(handlers.PostBarangHilangHandler(config.DB)))
	http.HandleFunc("/api/barang/ditemukan", middleware.RequireAuth(handlers.PostBarangDitemukanHandler(config.DB)))
	http.HandleFunc("/api/products", handlers.HandlerProductList)
}