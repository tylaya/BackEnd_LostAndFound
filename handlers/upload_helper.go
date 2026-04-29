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
	lokasiText := r.FormValue("lokasi")
	tanggalLaporan := r.FormValue("tanggal_laporan")
 
	if namaBarang == "" {
		http.Error(w, "Nama barang wajib diisi", http.StatusBadRequest)
		return
	}
 
	// Cari lokasi_id dari tabel lokasi berdasarkan nama
	var lokasiID int
	err = db.QueryRow(`SELECT id FROM lokasi WHERE nama_lokasi = ? LIMIT 1`, lokasiText).Scan(&lokasiID)
	if err != nil {
		// Tidak ditemukan, pakai ID pertama sebagai fallback
		err2 := db.QueryRow(`SELECT id FROM lokasi ORDER BY id LIMIT 1`).Scan(&lokasiID)
		if err2 != nil {
			lokasiID = 1
		}
	}
 
	if tanggalLaporan == "" {
		tanggalLaporan = time.Now().Format("2006-01-02")
	}
	if kategoriID == "" {
		kategoriID = "1"
	}
 
	// Handle foto (opsional)
	var destPath string
	file, header, fileErr := r.FormFile("foto")
	if fileErr == nil {
		defer file.Close()
		buf := make([]byte, 512)
		if _, err = file.Read(buf); err != nil {
			http.Error(w, "Gagal membaca file", http.StatusBadRequest)
			return
		}
		ft := http.DetectContentType(buf)
		if ft != "image/jpeg" && ft != "image/png" {
			http.Error(w, "Hanya JPEG dan PNG yang diizinkan.", http.StatusBadRequest)
			return
		}
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			http.Error(w, "Gagal memproses file", http.StatusInternalServerError)
			return
		}
		ext := filepath.Ext(header.Filename)
		now := time.Now().Format("20060102_150405")
		fname := fmt.Sprintf("barang_%s_%s%s", status, now, ext)
		destPath = "uploads/" + fname
 
		dst, createErr := os.Create(destPath)
		if createErr != nil {
			http.Error(w, "Gagal membuat file di server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err = io.Copy(dst, file); err != nil {
			http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
			return
		}
	}
 
	// INSERT barangs — schema baru: tidak ada kolom foto/lokasi, pakai lokasi_id
	result, err := db.Exec(
		`INSERT INTO barangs (user_id, nama_barang, deskripsi, kategori_id, status, lokasi_id, tanggal_laporan, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		userID, namaBarang, deskripsi, kategoriID, status, lokasiID, tanggalLaporan,
	)
	if err != nil {
		if destPath != "" {
			os.Remove(destPath)
		}
		http.Error(w, "Gagal menyimpan data ke database: "+err.Error(), http.StatusInternalServerError)
		return
	}
 
	newID, _ := result.LastInsertId()
 
	// Simpan foto ke tabel barang_fotos jika ada
	if destPath != "" && newID > 0 {
		_, _ = db.Exec(
			`INSERT INTO barang_fotos (barang_id, foto_path, created_at) VALUES (?, ?, NOW())`,
			newID, destPath,
		)
	}
 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  http.StatusCreated,
		Message: fmt.Sprintf("Laporan barang %s berhasil diunggah", status),
		Data: map[string]interface{}{
			"id":          newID,
			"nama_barang": namaBarang,
			"status":      status,
			"foto_path":   strings.ReplaceAll(destPath, "\\", "/"),
		},
	})
}