	package handlers

	import (
		"database/sql"
		"net/http"
	)

	// PostBarangHilangHandler menangani request dari halaman Laporkan Barang Hilang
	func PostBarangHilangHandler(db *sql.DB) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Memanggil fungsi dari upload_helper.go
			processUpload(db, w, r, "hilang")
		}
	}