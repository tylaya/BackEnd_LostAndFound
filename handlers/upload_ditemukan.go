		// handlers/upload_ditemukan.go
		package handlers

		import (
			"database/sql"
			"net/http"
		)


		func PostBarangDitemukanHandler(db *sql.DB) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				processUpload(db, w, r, "ditemukan")
			}
		}