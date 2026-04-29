// handlers/myitemshandler.go
package handlers

import (
	"backend-lostfound/config"
	"encoding/json"
	"net/http"
	"strings"
)

func GetMyItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	rows, err := config.DB.Query(`
		SELECT 
			b.id,
			b.user_id,
			b.nama_barang,
			b.deskripsi,
			b.status,
			CASE 
				WHEN COALESCE(
					(SELECT bf.foto_path FROM barang_fotos bf WHERE bf.barang_id = b.id ORDER BY bf.id ASC LIMIT 1),
					''
				) LIKE 'uploads/barang_hilang_%' THEN 'hilang'
				ELSE 'ditemukan'
			END AS tipe_laporan,
			COALESCE(l.nama_lokasi, '') AS lokasi,
			DATE_FORMAT(b.tanggal_laporan, '%Y-%m-%dT%H:%i:%sZ') AS tanggal_laporan,
			COALESCE(
				(SELECT bf.foto_path FROM barang_fotos bf WHERE bf.barang_id = b.id ORDER BY bf.id ASC LIMIT 1),
				''
			) AS foto
		FROM barangs b
		LEFT JOIN lokasi l ON l.id = b.lokasi_id
		WHERE b.user_id = ?
		ORDER BY b.created_at DESC
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
			&b.Status, &b.TipeLaporan, &b.Lokasi, &b.TanggalLaporan, &b.Foto,
		)
		if err != nil {
			continue
		}
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