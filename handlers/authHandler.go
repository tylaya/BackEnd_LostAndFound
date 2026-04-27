package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"backend-lostfound/config"
	"backend-lostfound/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// --- 1. REGISTER ---
// --- 1. REGISTER ---
func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"error": "Data tidak valid"}`, http.StatusBadRequest)
		return
	}

	// 🚨 FILTER EMAIL KAMPUS (UNNLAB) 🚨
	// Mengecek apakah email berakhiran "@student.unklab.ac.id"
	if !strings.HasSuffix(user.Email, "@student.unklab.ac.id") {
		http.Error(w, `{"error": "Akses ditolak! Hanya email @student.unklab.ac.id yang diizinkan untuk mendaftar."}`, http.StatusForbidden)
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	query := `INSERT INTO users (nama_depan, nama_belakang, email, password, nim, no_whatsapp, nomor_registrasi) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := config.DB.Exec(query, user.NamaDepan, user.NamaBelakang, user.Email, hashedPassword, user.NIM, user.NoWhatsapp, user.NomorRegistrasi)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			http.Error(w, `{"error": "Email atau NIM sudah terdaftar!"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "Gagal menyimpan data"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Registrasi berhasil!"}`))
}

// --- 2. LOGIN ---
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqUser models.User
	var dbUser models.User

	json.NewDecoder(r.Body).Decode(&reqUser)

	// 🚨 FILTER EMAIL KAMPUS (UNNLAB) 🚨
	// Mencegah proses query database jika email bukan email kampus
	if !strings.HasSuffix(reqUser.Email, "@student.unklab.ac.id") {
		http.Error(w, `{"error": "Akses ditolak! Gunakan email @student.unklab.ac.id untuk login."}`, http.StatusForbidden)
		return
	}

	query := `SELECT id, password, nama_depan FROM users WHERE email = ?`
	err := config.DB.QueryRow(query, reqUser.Email).Scan(&dbUser.ID, &dbUser.Password, &dbUser.NamaDepan)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Email tidak ditemukan"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(reqUser.Password)); err != nil {
		http.Error(w, `{"error": "Password salah"}`, http.StatusUnauthorized)
		return
	}

	// Buat Token dengan memasukkan ID user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": dbUser.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login berhasil, Halo " + dbUser.NamaDepan,
		"token":   tokenString,
	})
}

// --- 3. GET PROFILE ---
func Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ambil user_id dari context middleware
	userID := r.Context().Value("user_id").(int)

	var user models.User
	query := `SELECT id, nama_depan, nama_belakang, email, nim, no_whatsapp, nomor_registrasi FROM users WHERE id = ?`
	err := config.DB.QueryRow(query, int(userID)).Scan(
		&user.ID, &user.NamaDepan, &user.NamaBelakang, &user.Email, &user.NIM, &user.NoWhatsapp, &user.NomorRegistrasi,
	)

	if err != nil {
		http.Error(w, `{"error": "User tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// --- 4. CHANGE PASSWORD ---
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value("user_id").(int)

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var dbPassword string
	config.DB.QueryRow(`SELECT password FROM users WHERE id = ?`, userID).Scan(&dbPassword)

	// Cek password lama
	if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(req.OldPassword)); err != nil {
		http.Error(w, `{"error": "Password lama salah"}`, http.StatusUnauthorized)
		return
	}

	// Hash dan simpan password baru
	hashedNewPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	config.DB.Exec(`UPDATE users SET password = ? WHERE id = ?`, hashedNewPassword, userID)

	w.Write([]byte(`{"message": "Password berhasil diubah!"}`))
}

// --- 5. UPDATE STATUS BARANG ---

func UpdateStatusBarang(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// 1. Ambil ID User yang sedang login (dari token)

	userID := r.Context().Value("user_id").(int)

	// Format JSON Request: {"barang_id": 1, "status": "ditemukan"}

	var req struct {
		BarangID int `json:"barang_id"`

		Status string `json:"status"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	if req.Status != "hilang" && req.Status != "ditemukan" && req.Status != "selesai" {

		http.Error(w, `{"error": "Status tidak valid"}`, http.StatusBadRequest)

		return

	}

	// 2. QUERY DIAMANKAN: Cek id barang DAN user_id pemiliknya

	result, err := config.DB.Exec(`UPDATE barangs SET status = ? WHERE id = ? AND user_id = ?`, req.Status, req.BarangID, userID)

	if err != nil {

		http.Error(w, `{"error": "Gagal update status"}`, http.StatusInternalServerError)

		return

	}

	// 3. CEK APAKAH ADA BARANG YANG BERUBAH

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {

		http.Error(w, `{"error": "Akses Ditolak! Anda bukan pemilik barang ini atau barang tidak ditemukan."}`, http.StatusForbidden)

		return

	}

	w.Write([]byte(`{"message": "Status barang berhasil diupdate!"}`))

}
