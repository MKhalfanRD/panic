package seed

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"panic/internal/domain"
)

const (
	adminEmail  = "admin@panic.test"
	sellerEmail = "seller@panic.test"
	storeSlug   = "toko-contoh"
	productSlug = "paket-belajar"
)

func Ensure(db *gorm.DB, password string) error {
	var count int64
	if err := db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	password = strings.TrimSpace(password)
	if password == "" {
		password = "panic-dev"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := domain.User{Name: "Admin", Email: adminEmail, PasswordHash: string(hash), Role: domain.RoleAdmin}
	sellerUser := domain.User{Name: "Penjual Contoh", Email: sellerEmail, PasswordHash: string(hash), Role: domain.RoleSeller}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	if err := db.Create(&sellerUser).Error; err != nil {
		return err
	}
	seller := domain.Seller{
		UserID:      sellerUser.ID,
		StoreName:   "Toko Contoh",
		Slug:        storeSlug,
		Description: "Contoh toko produk digital. Ubah teks ini dari data seller, atau tambah produk baru lewat admin.",
	}
	if domain.ReservedSlug(seller.Slug) {
		return errors.New("slug toko contoh bentrok dengan jalur cadangan")
	}
	if err := db.Create(&seller).Error; err != nil {
		return err
	}
	category := domain.Category{Name: "Pendidikan", Slug: "pendidikan"}
	if err := db.Create(&category).Error; err != nil {
		return err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	ends := time.Now().In(loc).Add(7 * 24 * time.Hour).Truncate(time.Minute)
	product := domain.Product{
		SellerID:      seller.ID,
		CategoryID:    category.ID,
		Name:          "Paket Belajar Riang",
		Slug:          productSlug,
		Price:         99000,
		OriginalPrice: 149000,
		Status:        domain.ProductPublished,
		Landing:       sampleLanding(&ends),
	}
	cleaned, err := domain.ValidateProduct(product)
	if err != nil {
		return err
	}
	return db.Omit("Seller", "Category").Create(&cleaned).Error
}

func sampleLanding(ends *time.Time) domain.Landing {
	return domain.Landing{
		MetaTitle:       "Paket Belajar Riang — 48 video singkat",
		MetaDescription: "Video pendamping belajar untuk orang tua. Bayar sekali, tonton kapan saja.",
		HeroHeadline:    "Belajar singkat, tanpa layar seharian",
		HeroSubhead:     "48 video pendamping untuk orang tua. Bayar sekali, putar kapan saja di rumah.",
		Stats: []domain.Stat{
			{Value: "48", Label: "Video"},
			{Value: "6", Label: "Topik"},
			{Value: "15 mnt", Label: "Durasi"},
			{Value: "1x", Label: "Bayar"},
		},
		Problems: []string{
			"Anak cepat bosan dengan penjelasan yang panjang",
			"Orang tua kehabisan ide kegiatan belajar",
			"Materi gratis tersebar dan sulit dipilih",
		},
		ProblemClose: "Paket ini merangkum kegiatan singkat yang bisa diulang sendiri.",
		Includes: []string{
			"Video pengenalan huruf dan angka",
			"Lembar latihan yang bisa dicetak",
			"Ide permainan 10 menit",
			"Panduan pendampingan orang tua",
		},
		IncludeNote: "Status akses tercantum di halaman pembayaran setelah pesanan lunas.",
		Benefits: []string{
			"Satu pembayaran, tanpa langganan bulanan",
			"Bisa diulang sesuai kecepatan anak",
			"Kegiatan singkat, mudah disisipkan dalam hari",
		},
		Bonuses: []domain.Bonus{
			{Name: "Lembar cetak 30 halaman", Value: "Senilai Rp 49.000"},
			{Name: "Checklist mingguan", Value: "Senilai Rp 19.000"},
		},
		Testimonials: []domain.Testimonial{
			{Quote: "Anak saya mau mengulang video huruf tanpa diminta.", Author: "Rani, Depok", Stars: 5},
			{Quote: "Durasi singkat jadi tidak berebut layar.", Author: "Dimas, Surabaya", Stars: 5},
		},
		UrgencyText: "Harga perkenalan berlaku sampai waktu di bawah ini",
		OfferEndsAt: ends,
		Checklist: []string{
			"48 video pendek",
			"Lembar latihan dan bonus cetak",
			"Pesanan tercatat beserta status bayar",
		},
		CTALabel: "Beli sekarang",
		FAQ: []domain.FAQ{
			{Question: "Untuk siapa paket ini?", Answer: "Orang tua yang mendampingi anak usia dini di rumah."},
			{Question: "Kapan akses dikirim?", Answer: "Pesanan berstatus lunas tercatat di halaman status. Pengiriman berkas lewat email menyusul pada pembaruan berikutnya."},
			{Question: "Apakah ini langganan?", Answer: "Tidak. Satu kali bayar untuk paket ini."},
		},
	}
}
