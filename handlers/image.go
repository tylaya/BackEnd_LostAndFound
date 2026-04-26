package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
)

// imageURLToBase64 membaca file gambar dan mengubahnya menjadi string Base64.
func imageURLToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// GetImageHandler mengambil path foto dari database dan mengembalikan Base64-nya
// (Diubah menggunakan Dependency Injection db *sql.DB agar tidak terjadi memory leak)
func GetImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Menggunakan Method GET karena kita mengambil data, bukan POST
		if r.Method != http.MethodGet {
			http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
			return
		}

		// Mengambil ID Barang dari URL Query (misal: /api/image?id_barang=1)
		idBarang := r.URL.Query().Get("id_barang")
		if idBarang == "" {
			http.Error(w, "ID Barang diperlukan", http.StatusBadRequest)
			return
		}

		var pathFoto string
		
		err := db.QueryRow(`SELECT foto FROM Barangs WHERE id = ?`, idBarang).Scan(&pathFoto)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Barang tidak ditemukan", http.StatusNotFound)
			} else {
				http.Error(w, "Terjadi kesalahan pada database", http.StatusInternalServerError)
			}
			return
		}

		// Convert file yang path-nya ditemukan ke Base64 
		base64Str, err := imageURLToBase64(pathFoto)
		if err != nil {
			http.Error(w, "Gagal memuat file gambar dari server", http.StatusInternalServerError)
			return  
		}

		// Kirim Response dalam bentuk JSON agar konsisten dengan API lainnya
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
			"image_base64": base64Str,
		})
	}
}