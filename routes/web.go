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

	// Rute Serve Static Files (foto barang dari folder uploads/)
	// Contoh akses: http://localhost:8081/uploads/barang_hilang_20260427_123055.jpeg
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// Rute Ambil Gambar (legacy endpoint, tetap dipertahankan)
	http.HandleFunc("/api/image", handlers.GetImageHandler(config.DB))

	// Rute VIP (Wajib Token JWT)
	http.HandleFunc("/api/profile", middleware.RequireAuth(handlers.Profile))
	http.HandleFunc("/api/change-password", middleware.RequireAuth(handlers.ChangePassword))
	http.HandleFunc("/api/status-barang", middleware.RequireAuth(handlers.UpdateStatusBarang))
	
	http.HandleFunc("/api/barang/hilang", middleware.RequireAuth(handlers.PostBarangHilangHandler(config.DB)))
	http.HandleFunc("/api/barang/ditemukan", middleware.RequireAuth(handlers.PostBarangDitemukanHandler(config.DB)))
	http.HandleFunc("/api/products", handlers.HandlerProductList)

	// Rute VIP: ambil barang milik user yang sedang login
	http.HandleFunc("/api/my-items", middleware.RequireAuth(handlers.GetMyItemsHandler))
}