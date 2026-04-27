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
            // id asli dri token dan pastikan data float64
            ctx := context.WithValue(r.Context(), "user_id", claims["user_id"].(float64)) 

            next(w, r.WithContext(ctx))
        } else {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(`{"error": "Gagal membaca isi token!"}`))
            return
        }
	}
}
