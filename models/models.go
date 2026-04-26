package models

// Struct untuk tabel users
type User struct {
	ID              int    `json:"id"`
	NamaDepan       string `json:"nama_depan"`
	NamaBelakang    string `json:"nama_belakang"`
	Email           string `json:"email"`
	Password        string `json:"password,omitempty"` // omitempty agar tidak bocor saat dikirim ke frontend
	NIM             string `json:"nim"`
	NoWhatsapp      string `json:"no_whatsapp"`
	NomorRegistrasi string `json:"nomor_registrasi"`
}

// Struct untuk tabel barangs
type Barang struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	NamaBarang     string `json:"nama_barang"`
	Deskripsi      string `json:"deskripsi"`
	KategoriID     int    `json:"kategori_id"`
	Status         string `json:"status"` // 'hilang', 'ditemukan', 'selesai'
	Lokasi         string `json:"lokasi"`
	TanggalLaporan string `json:"tanggal_laporan"`
	Foto           string `json:"foto"`
}

// Struct untuk tabel categories
type Category struct {
	ID           int    `json:"id"`
	NamaKategori string `json:"nama_kategori"`
}
