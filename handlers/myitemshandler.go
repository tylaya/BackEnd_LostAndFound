// handlers/myitemshandler.go
package handlers

import (
	"backend-lostfound/config"
	"encoding/json"
	"net/http"
	"strings"
)

// GetMyItemsHandler mengambil daftar barang milik user yang sedang login (berdasarkan JWT)
func GetMyItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	// Ambil user_id dari context (sudah diset oleh middleware RequireAuth)
	userID := r.Context().Value("user_id").(int)

	rows, err := config.DB.Query(`
		SELECT id, user_id, nama_barang, deskripsi, status, lokasi, tanggal_laporan, foto 
		FROM barangs 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
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
		// Fix path Windows backslash → forward slash
		b.Foto = strings.ReplaceAll(b.Foto, "\\", "/")
		listBarang = append(listBarang, b)
	}

	if listBarang == nil {
		listBarang = []BarangList{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "Berhasil mengambil daftar barang milik user",
		"data":    listBarang,
	})
}
