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
	UserID         int    `json:"user_id"`
	NamaBarang     string `json:"nama_barang"`
	Deskripsi      string `json:"deskripsi"`
	Status         string `json:"status"`
	TipeLaporan    string `json:"tipe_laporan"` // 'hilang' atau 'ditemukan' (tipe awal laporan)
	Lokasi         string `json:"lokasi"`
	TanggalLaporan string `json:"tanggal_laporan"`
	Foto           string `json:"foto"`
}

func HandlerProductList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// JOIN dengan tabel lokasi untuk dapat nama lokasi
	// LEFT JOIN dengan barang_fotos untuk dapat path foto (ambil satu foto pertama saja)
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
		ORDER BY b.created_at DESC
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
			&b.Status, &b.TipeLaporan, &b.Lokasi, &b.TanggalLaporan, &b.Foto,
		)
		if err != nil {
			continue
		}
		// Normalisasi path: backslash → forward slash
		b.Foto = strings.ReplaceAll(b.Foto, "\\", "/")
		listBarang = append(listBarang, b)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "Berhasil mengambil daftar barang",
		"data":    listBarang,
	})
}