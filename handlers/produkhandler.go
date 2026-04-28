// handlers/produkhandler.go
package handlers

import (
	"backend-lostfound/config"
	"encoding/json"
	"net/http"
	"strings"
)

type BarangList struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`       // ← TAMBAHAN: untuk filter "Barang Saya"
	NamaBarang     string `json:"nama_barang"`
	Deskripsi      string `json:"deskripsi"`
	Status         string `json:"status"`
	Lokasi         string `json:"lokasi"`
	TanggalLaporan string `json:"tanggal_laporan"`
	Foto           string `json:"foto"`
}

// HandlerProductList untuk mengambil semua data barang
func HandlerProductList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Tambah user_id di SELECT
	rows, err := config.DB.Query(`
		SELECT id, user_id, nama_barang, deskripsi, status, lokasi, tanggal_laporan, foto 
		FROM barangs 
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil data database"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listBarang []BarangList

	for rows.Next() {
		var b BarangList
		err := rows.Scan(
			&b.ID, &b.UserID, &b.NamaBarang, &b.Deskripsi,
			&b.Status, &b.Lokasi, &b.TanggalLaporan, &b.Foto,
		)
		if err != nil {
			continue
		}
		// Fix path Windows backslash → forward slash agar bisa jadi URL
		b.Foto = strings.ReplaceAll(b.Foto, "\\", "/")
		listBarang = append(listBarang, b)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "Berhasil mengambil daftar barang",
		"data":    listBarang,
	})
}