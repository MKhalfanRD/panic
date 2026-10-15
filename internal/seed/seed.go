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
	adminEmail        = "admin@panic.test"
	sellerEmail       = "seller@panic.test"
	storeSlug         = "nota-rapi"
	productSlug       = "paket-invoice"
	taxSlug           = "kit-pajak-seller"
	legacyStoreSlug   = "toko-contoh"
	legacyProductSlug = "paket-belajar"
)

func Ensure(db *gorm.DB, password string) error {
	var count int64
	if err := db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		if err := upgradeDummy(db); err != nil {
			return err
		}
		return ensureTaxKit(db)
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
	sellerUser := domain.User{Name: "Nota Rapi", Email: sellerEmail, PasswordHash: string(hash), Role: domain.RoleSeller}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	if err := db.Create(&sellerUser).Error; err != nil {
		return err
	}
	seller := domain.Seller{
		UserID:      sellerUser.ID,
		StoreName:   "Nota Rapi",
		Slug:        storeSlug,
		Description: "Dokumen siap pakai untuk freelancer, jasa kecil, dan seller online.",
	}
	if domain.ReservedSlug(seller.Slug) {
		return errors.New("slug toko bentrok dengan jalur cadangan")
	}
	if err := db.Create(&seller).Error; err != nil {
		return err
	}
	category := domain.Category{Name: "Bisnis", Slug: "bisnis"}
	if err := db.Create(&category).Error; err != nil {
		return err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	ends := time.Now().In(loc).Add(7 * 24 * time.Hour).Truncate(time.Minute)
	product := sampleProduct(seller.ID, category.ID, &ends)
	cleaned, err := domain.ValidateProduct(product)
	if err != nil {
		return err
	}
	if err := db.Omit("Seller", "Category").Create(&cleaned).Error; err != nil {
		return err
	}
	return ensureTaxKit(db)
}

func sampleProduct(sellerID, categoryID uint, ends *time.Time) domain.Product {
	return domain.Product{
		SellerID:      sellerID,
		CategoryID:    categoryID,
		Name:          "Paket Invoice Freelancer",
		Slug:          productSlug,
		Price:         49000,
		OriginalPrice: 89000,
		Status:        domain.ProductPublished,
		Landing:       sampleLanding(ends),
	}
}

func sampleLanding(ends *time.Time) domain.Landing {
	return domain.Landing{
		MetaTitle:       "Paket Invoice Freelancer — siap tagih klien",
		MetaDescription: "Invoice, kwitansi, dan surat penawaran siap isi. Bayar sekali, pakai untuk setiap klien.",
		HeroHeadline:    "Tagih klien tanpa membuat invoice dari nol",
		HeroSubhead:     "Empat dokumen untuk freelancer dan jasa kecil. Isi nama, jasa, dan nominal, lalu kirim.",
		Stats: []domain.Stat{
			{Value: "4", Label: "Dokumen"},
			{Value: "1x", Label: "Bayar"},
			{Value: "Siap salin", Label: "Format"},
			{Value: "Semua jasa", Label: "Cocok untuk"},
		},
		Problems: []string{
			"Invoice tiap klien dibuat ulang, formatnya tidak sama",
			"Tagihan tidak menyebut jatuh tempo, klien menunda bayar",
			"Penawaran hanya lewat chat, lalu sulit dijadikan bukti",
		},
		ProblemClose: "Isi sekali, pakai ulang. Setiap klien menerima dokumen yang sama rapi.",
		Includes: []string{
			"Template invoice",
			"Template kwitansi",
			"Surat penawaran satu halaman",
			"Checklist data klien sebelum mulai kerja",
		},
		IncludeNote: "Berkas belum terkirim otomatis. Setelah lunas, status tercatat di halaman pembayaran dan penjual mengirim dokumen ke email Anda.",
		Benefits: []string{
			"Satu format untuk semua klien",
			"Ada tempat untuk nominal, jatuh tempo, dan rekening",
			"Bisa disalin untuk proyek berikutnya",
			"Bayar sekali, tanpa langganan",
		},
		Bonuses: []domain.Bonus{
			{Name: "Contoh terisi untuk jasa desain", Value: "Termasuk"},
			{Name: "Contoh terisi untuk les dan tulis", Value: "Termasuk"},
		},
		UrgencyText: "Harga perkenalan berlaku sampai waktu di bawah ini",
		OfferEndsAt: ends,
		Checklist: []string{
			"Invoice, kwitansi, dan surat penawaran",
			"Checklist data klien",
			"Dua contoh dokumen yang sudah terisi",
		},
		CTALabel: "Ambil paketnya",
		FAQ: []domain.FAQ{
			{Question: "Untuk siapa paket ini?", Answer: "Freelancer, jasa rumahan, dan usaha kecil yang menagih per proyek."},
			{Question: "Format berkasnya apa?", Answer: "Dokumen yang bisa disalin di Google Docs atau Word."},
			{Question: "Kapan berkas dikirim?", Answer: "Setelah pesanan lunas. Pengiriman otomatis belum tersedia, jadi penjual mengirim berkas ke email pembeli."},
			{Question: "Apakah ini langganan?", Answer: "Tidak. Satu kali bayar, dipakai ulang untuk klien berikutnya."},
		},
	}
}

// upgradeDummy replaces the original sample course when that row is still in the database.
func upgradeDummy(db *gorm.DB) error {
	var product domain.Product
	err := db.Where("slug = ?", legacyProductSlug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	var seller domain.Seller
	if err := db.First(&seller, product.SellerID).Error; err != nil {
		return err
	}
	if seller.Slug == legacyStoreSlug {
		seller.StoreName = "Nota Rapi"
		seller.Slug = storeSlug
		seller.Description = "Dokumen siap pakai untuk freelancer, jasa kecil, dan seller online."
		if err := db.Model(&seller).Select("StoreName", "Slug", "Description").Updates(&seller).Error; err != nil {
			return err
		}
		if err := db.Model(&domain.User{}).Where("id = ? AND name = ?", seller.UserID, "Penjual Contoh").Update("name", "Nota Rapi").Error; err != nil {
			return err
		}
	}

	var category domain.Category
	if err := db.First(&category, product.CategoryID).Error; err != nil {
		return err
	}
	if category.Slug == "pendidikan" {
		category.Name = "Bisnis"
		category.Slug = "bisnis"
		if err := db.Model(&category).Select("Name", "Slug").Updates(&category).Error; err != nil {
			return err
		}
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	ends := time.Now().In(loc).Add(7 * 24 * time.Hour).Truncate(time.Minute)
	next := sampleProduct(product.SellerID, product.CategoryID, &ends)
	next.ID = product.ID
	cleaned, err := domain.ValidateProduct(next)
	if err != nil {
		return err
	}
	return db.Model(&product).Select("Name", "Slug", "Price", "OriginalPrice", "Status", "Landing").Updates(&cleaned).Error
}

func ensureTaxKit(db *gorm.DB) error {
	var seller domain.Seller
	err := db.Where("slug = ?", storeSlug).First(&seller).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if seller.Description == "Dokumen tagihan siap pakai untuk freelancer dan jasa kecil." {
		if err := db.Model(&seller).Update("description", "Dokumen siap pakai untuk freelancer, jasa kecil, dan seller online.").Error; err != nil {
			return err
		}
	}

	var existing int64
	if err := db.Model(&domain.Product{}).Where("seller_id = ? AND slug = ?", seller.ID, taxSlug).Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	var category domain.Category
	err = db.Where("slug = ?", "bisnis").First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		category = domain.Category{Name: "Bisnis", Slug: "bisnis"}
		if err := db.Create(&category).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	product := taxProduct(seller.ID, category.ID)
	cleaned, err := domain.ValidateProduct(product)
	if err != nil {
		return err
	}
	return db.Omit("Seller", "Category").Create(&cleaned).Error
}

func taxProduct(sellerID, categoryID uint) domain.Product {
	return domain.Product{
		SellerID:      sellerID,
		CategoryID:    categoryID,
		Name:          "Kit Pajak Seller",
		Slug:          taxSlug,
		Price:         79000,
		OriginalPrice: 129000,
		Status:        domain.ProductPublished,
		Landing:       taxLanding(),
	}
}

func taxLanding() domain.Landing {
	return domain.Landing{
		MetaTitle:       "Kit Pajak Seller — catat omzet toko online",
		MetaDescription: "Excel untuk omzet, surat pernyataan, checklist SKB, dan bukti pungut marketplace. Bayar sekali.",
		HeroHeadline:    "Catat omzet toko online sebelum kena potong",
		HeroSubhead:     "Empat lembar Excel untuk seller Shopee, Tokopedia, Lazada, dan Blibli. Isi angka, sisanya terhitung sendiri.",
		Stats: []domain.Stat{
			{Value: "4", Label: "Lembar"},
			{Value: "2026", Label: "Tahun"},
			{Value: "Excel", Label: "Format"},
			{Value: "1x", Label: "Bayar"},
		},
		Problems: []string{
			"Omzet tersebar di banyak marketplace, sulit dijumlahkan",
			"Surat pernyataan omzet ditulis ulang untuk tiap marketplace",
			"Bukti potong pajak tercecer dan tidak tercatat",
		},
		ProblemClose: "Satu file untuk omzet, surat, checklist, dan bukti pungut.",
		Includes: []string{
			"Lembar omzet 12 bulan lintas marketplace",
			"Draf surat pernyataan omzet yang tinggal dicetak",
			"Checklist surat dan SKB",
			"Catatan bukti pungut pajak marketplace",
		},
		IncludeNote: "Alat bantu hitung, bukan nasihat pajak. Aturan bisa berubah. Berkas dikirim ke email setelah pesanan lunas.",
		Benefits: []string{
			"Total omzet dan sisa batas terhitung sendiri",
			"Status aman atau mendekati batas berubah sendiri",
			"Surat terbentuk setelah tujuh isian",
			"Potongan pajak dicek terhadap tarif 0,5%",
		},
		Bonuses: []domain.Bonus{
			{Name: "Contoh angka yang bisa ditimpa", Value: "Termasuk"},
			{Name: "Catatan sumber aturan di tiap lembar", Value: "Termasuk"},
		},
		UrgencyText: "Pemungutan marketplace berlaku sejak 1 Agustus 2026 bagi yang tidak dikecualikan",
		Checklist: []string{
			"Lembar omzet, surat, checklist, dan bukti pungut",
			"Contoh isian yang bisa ditimpa",
			"Bukan pengganti konsultasi pajak",
		},
		CTALabel: "Ambil kitnya",
		FAQ: []domain.FAQ{
			{Question: "Untuk siapa kit ini?", Answer: "Seller marketplace yang ingin memantau omzet dan menyiapkan surat pernyataan."},
			{Question: "Format berkasnya apa?", Answer: "Satu file Excel dengan empat lembar: Omzet Saya, Surat Pernyataan, Checklist SKB, dan Bukti Pungut."},
			{Question: "Apakah ini nasihat pajak?", Answer: "Tidak. Ini alat bantu hitung. Surat di dalamnya draf contoh. Jika marketplace punya format resmi, pakai format itu. Cocokkan aturan terbaru dengan DJP atau konsultan pajak."},
			{Question: "Kapan berkas dikirim?", Answer: "Setelah pesanan lunas. Pengiriman otomatis belum tersedia, jadi penjual mengirim file Excel ke email pembeli."},
		},
	}
}
