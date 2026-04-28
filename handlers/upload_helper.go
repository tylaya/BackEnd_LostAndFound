// handlers/upload_helper.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	err := os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		fmt.Println("Gagal membuat direktori uploads:", err)
	}
}

type APIResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func processUpload(db *sql.DB, w http.ResponseWriter, r *http.Request, status string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Ukuran file terlalu besar, maksimal 10MB", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)
	namaBarang := r.FormValue("nama_barang")
	deskripsi := r.FormValue("deskripsi")
	kategoriID := r.FormValue("kategori_id")
	lokasi := r.FormValue("lokasi")
	tanggalLaporan := r.FormValue("tanggal_laporan")

	if namaBarang == "" {
		http.Error(w, "Nama barang wajib diisi", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("foto")
	if err != nil {
		http.Error(w, "Gagal membaca file foto dari request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err = file.Read(buffer); err != nil {
		http.Error(w, "Gagal membaca file", http.StatusBadRequest)
		return
	}

	fileType := http.DetectContentType(buffer)
	if fileType != "image/jpeg" && fileType != "image/png" {
		http.Error(w, "Invalid file type. Hanya JPEG dan PNG yang diizinkan.", http.StatusBadRequest)
		return
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Gagal memproses file setelah validasi", http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(header.Filename)
	now := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("barang_%s_%s%s", status, now, ext)
	// Gunakan forward slash agar path konsisten di semua OS
	destPath := "uploads/" + filename

	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Gagal membuat file di server", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO Barangs (user_id, nama_barang, deskripsi, kategori_id, status, lokasi, tanggal_laporan, foto, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := db.Exec(query, userID, namaBarang, deskripsi, kategoriID, status, lokasi, tanggalLaporan, destPath)
	if err != nil {
		os.Remove(destPath)
		http.Error(w, "Gagal menyimpan data ke database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ← TAMBAHAN: ambil ID yang baru dibuat untuk dikirim ke frontend
	newID, err := result.LastInsertId()
	if err != nil {
		newID = 0
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(APIResponse{
		Status:  http.StatusCreated,
		Message: fmt.Sprintf("Laporan barang %s berhasil diunggah", status),
		Data: map[string]interface{}{
			"id":          newID,                                   // ← ID asli dari DB
			"nama_barang": namaBarang,
			"status":      status,
			"foto_path":   strings.ReplaceAll(destPath, "\\", "/"), // forward slash
		},
	})
}