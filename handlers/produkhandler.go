package handlers

import (
	"backend-lostfound/config"
	"encoding/json"
	"net/http"
)

type BarangList struct {
    ID             int    `json:"id"`
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

    // Query ke tabel barangs (pakai 's' sesuai database Anda)
    rows, err := config.DB.Query(`SELECT id, nama_barang, deskripsi, status, lokasi, tanggal_laporan, foto FROM barangs ORDER BY created_at DESC`)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengambil data database"}`, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var listBarang []BarangList

    for rows.Next() {
        var b BarangList
        // Scan data dari baris database ke struct
        err := rows.Scan(&b.ID, &b.NamaBarang, &b.Deskripsi, &b.Status, &b.Lokasi, &b.TanggalLaporan, &b.Foto)
        if err != nil {
            continue // Lewati jika ada satu baris yang error
        }
        listBarang = append(listBarang, b)
    }

    // Kirim response sukses
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":  200,
        "message": "Berhasil mengambil daftar barang",
        "data":    listBarang,
    })
}