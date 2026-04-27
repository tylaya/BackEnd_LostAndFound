package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth adalah fungsi middleware untuk melindungi rute yang butuh login
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Ambil header Authorization dari request Frontend
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Akses ditolak. Token tidak ditemukan!"}`))
			return
		}

		// 2. Pisahkan kata "Bearer" dan isi tokennya
		// Format yang benar: "Bearer eyJhbGciOiJIUzI1NiIsIn..."
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Format token tidak valid! Harus menggunakan Bearer."}`))
			return
		}

		tokenString := tokenParts[1]

		// 3. Baca dan verifikasi Token menggunakan JWT_SECRET yang ada di file .env
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan metode enkripsi tokennya benar (HMAC)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Jika token gagal dibaca, palsu, atau sudah kedaluwarsa
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Token tidak valid atau sudah kadaluarsa!"}`))
			return
		}

		// 4. Jika Token Valid, ekstrak data (claims) di dalamnya
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// JWT membaca angka sebagai float64 secara default.
			// Kita lakukan type assertion ke float64 terlebih dahulu.
			if userIDFloat, ok := claims["user_id"].(float64); ok {

				// Konversi float64 menjadi int (karena ID di database kamu adalah INT)
				userID := int(userIDFloat)

				// Masukkan userID (yang sekarang sudah dijamin bertipe int) ke dalam Context
				ctx := context.WithValue(r.Context(), "user_id", userID)

				// Izinkan request melanjutkan perjalanannya ke Handler
				next(w, r.WithContext(ctx))
			} else {
				// Jika gagal membaca user_id dari token
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Format token tidak valid: user_id gagal diekstrak!"}`))
				return
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Gagal membaca isi token!"}`))
			return
		}
	}
}
