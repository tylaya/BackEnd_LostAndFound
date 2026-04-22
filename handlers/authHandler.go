package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"backend-lostfound/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Struktur data yang diharapkan dari Frontend saat mendaftar
type RegisterRequest struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	StudentID string `json:"student_id"`
	Faculty   string `json:"faculty"` //ganti dengan jurusan
	Password  string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Format data tidak valid"}`))
		return
	}

	// --- 4 HAL PENTING: Poin D (Robust Validation) ---

	// 1. Validasi Email Mahasiswa
	if !strings.HasSuffix(req.Email, "@student.unklab.ac.id") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Gagal! Wajib menggunakan email @student.unklab.ac.id"}`))
		return
	}

	// 2. Sanitasi & Format Nomor WhatsApp (Ubah 08 menjadi 628)
	// Agar link wa.me/ di frontend bisa langsung bekerja
	if strings.HasPrefix(req.Phone, "08") {
		req.Phone = "628" + req.Phone[2:]
	} else if strings.HasPrefix(req.Phone, "+62") {
		req.Phone = "62" + req.Phone[3:]
	}

	// --- 4 HAL PENTING: Poin A (Keamanan Password / Bcrypt) ---

	// Hash password sebelum disimpan ke database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Gagal memproses password"}`))
		return
	}

	// SIMPAN KE DATABASE
	query := `INSERT INTO users (full_name, email, phone, student_id, faculty, password) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = config.DB.Exec(query, req.FullName, req.Email, req.Phone, req.StudentID, req.Faculty, string(hashedPassword))

	if err != nil {
		// Jika email atau student_id sudah ada (karena ada aturan UNIQUE di database)
		if strings.Contains(err.Error(), "Duplicate entry") {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error": "Email atau NIM sudah terdaftar"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Gagal mendaftarkan user"}`))
		return
	}

	// Response Sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Registrasi berhasil! Silakan login."}`))
}

// Struktur data untuk Login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Format data tidak valid"}`))
		return
	}

	// 1. Cari user berdasarkan email di database
	var id int
	var hashedPassword, role, fullName string

	query := `SELECT id, password, role, full_name FROM users WHERE email = ?`
	err = config.DB.QueryRow(query, req.Email).Scan(&id, &hashedPassword, &role, &fullName)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Email atau password salah"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Terjadi kesalahan pada server"}`))
		return
	}

	// 2. Cocokkan password yang diinput dengan password hash di database
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Email atau password salah"}`))
		return
	}

	// 3. Jika cocok, Buat JWT Token
	// Kita simpan ID dan Role di dalam token agar frontend tahu siapa yang login
	claims := jwt.MapClaims{
		"user_id":   id,
		"role":      role,
		"full_name": fullName,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := os.Getenv("JWT_SECRET") // Mengambil rahasia dari file .env
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Gagal membuat token"}`))
		return
	}

	// 4. Kirim respon sukses beserta Token
	response := map[string]string{
		"message": "Login berhasil",
		"token":   tokenString,
		"role":    role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func Profile(w http.ResponseWriter, r *http.Request) {
	// Karena sudah melewati "Satpam" (Middleware), kita yakin user ini valid
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"message": "Selamat datang di area VIP!",
		"info": "Jika Anda melihat pesan ini, berarti token JWT Anda valid dan Satpam mengizinkan Anda masuk."
	}`))
}
