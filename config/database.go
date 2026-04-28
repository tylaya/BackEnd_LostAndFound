// config/database.go
package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	// Ambil konfigurasi DSN dari file .env
	dsn := os.Getenv("DSN")

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("Gagal membuka koneksi database: " + err.Error())
	}

	// Cek apakah database benar-benar tersambung
	err = DB.Ping()
	if err != nil {
		panic("Database tidak merespon: " + err.Error())
	}

	fmt.Println("🚀 Koneksi Database MySQL Berhasil!")
}
