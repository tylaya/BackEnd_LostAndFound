package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth adalah "Satpam" kita
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Ambil token dari header "Authorization"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Akses ditolak! Token tidak ditemukan."}`))
			return
		}

		// 2. Format token yang benar adalah "Bearer <token_panjang>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Cek apakah token tersebut asli dan dibuat oleh server kita
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Token tidak valid atau sudah kadaluarsa!"}`))
			return
		}

		// Jika token aman, persilakan user masuk ke rute yang mereka tuju
		next(w, r)
	}
}
