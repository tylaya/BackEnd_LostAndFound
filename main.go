package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend-lostfound/config"
	"backend-lostfound/routes"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: File .env tidak ditemukan")
	}

	// 2. Koneksi ke Database
	config.ConnectDB()

	// 3. Setup Routes (Nanti kita isi di web.go)
	routes.RegisterRoutes()

	// 4. Setup CORS (Cross-Origin Resource Sharing)
	c := cors.New(cors.Options{
		// Tambahkan port 8082 ke dalam daftar yang diizinkan
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8082"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// 3. Bungkus rute bawaan Golang dengan CORS
	handler := c.Handler(http.DefaultServeMux)

	// 4. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8081"
	}

	fmt.Println("Server berjalan di http://localhost" + port)
	err = http.ListenAndServe(port, handler)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
