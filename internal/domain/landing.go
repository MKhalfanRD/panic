package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const maxPrice int64 = 1_000_000_000

type Stat struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Bonus struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Testimonial struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
	Stars  int    `json:"stars"`
}

type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type Landing struct {
	MetaTitle       string        `json:"meta_title"`
	MetaDescription string        `json:"meta_description"`
	HeroHeadline    string        `json:"hero_headline"`
	HeroSubhead     string        `json:"hero_subhead"`
	HeroImageURL    string        `json:"hero_image_url"`
	Stats           []Stat        `json:"stats"`
	Problems        []string      `json:"problems"`
	ProblemClose    string        `json:"problem_close"`
	Includes        []string      `json:"includes"`
	IncludeNote     string        `json:"include_note"`
	Benefits        []string      `json:"benefits"`
	Bonuses         []Bonus       `json:"bonuses"`
	Testimonials    []Testimonial `json:"testimonials"`
	UrgencyText     string        `json:"urgency_text"`
	OfferEndsAt     *time.Time    `json:"offer_ends_at,omitempty"`
	Checklist       []string      `json:"checklist"`
	CTALabel        string        `json:"cta_label"`
	FAQ             []FAQ         `json:"faq"`
}

func ValidateProduct(in Product) (Product, error) {
	var errs []string
	out := in
	out.Name = strings.TrimSpace(in.Name)
	out.Slug = strings.TrimSpace(in.Slug)
	out.Status = strings.TrimSpace(in.Status)
	out.ImageURL = strings.TrimSpace(in.ImageURL)

	if n := utf8.RuneCountInString(out.Name); n < 2 || n > 160 {
		errs = append(errs, "Nama produk harus 2–160 karakter")
	}
	if !ValidSlug(out.Slug) {
		errs = append(errs, "Slug produk harus huruf kecil, angka, dan tanda hubung")
	}
	if out.Status != ProductDraft && out.Status != ProductPublished {
		errs = append(errs, "Status harus draft atau tayang")
	}
	if out.Price < 0 || out.OriginalPrice < 0 || out.Price > maxPrice || out.OriginalPrice > maxPrice {
		errs = append(errs, "Harga tidak valid")
	}
	if out.Status == ProductPublished && out.Price <= 0 {
		errs = append(errs, "Harga jual wajib lebih dari 0 untuk produk tayang")
	}
	if out.OriginalPrice > 0 && out.OriginalPrice <= out.Price {
		errs = append(errs, "Harga coret harus lebih tinggi dari harga jual")
	}
	if out.SellerID == 0 {
		errs = append(errs, "Toko wajib dipilih")
	}
	if out.CategoryID == 0 {
		errs = append(errs, "Kategori wajib dipilih")
	}
	if out.ImageURL != "" {
		safe, ok := SafeImageURL(out.ImageURL)
		if !ok {
			errs = append(errs, "URL gambar produk harus http(s) atau path /...")
		} else {
			out.ImageURL = safe
		}
	}

	landing, landErrs := CleanLanding(in.Landing, out.Status)
	out.Landing = landing
	errs = append(errs, landErrs...)
	if len(errs) > 0 {
		return out, errors.New(strings.Join(unique(errs), ". "))
	}
	return out, nil
}

func CleanLanding(in Landing, status string) (Landing, []string) {
	var errs []string
	out := Landing{
		MetaTitle:       limit(in.MetaTitle, 70, "Judul SEO", &errs),
		MetaDescription: limit(in.MetaDescription, 180, "Deskripsi SEO", &errs),
		HeroHeadline:    limit(in.HeroHeadline, 160, "Headline", &errs),
		HeroSubhead:     limit(in.HeroSubhead, 400, "Subheadline", &errs),
		ProblemClose:    limit(in.ProblemClose, 240, "Kalimat penutup masalah", &errs),
		IncludeNote:     limit(in.IncludeNote, 240, "Catatan isi", &errs),
		UrgencyText:     limit(in.UrgencyText, 160, "Teks urgensi", &errs),
		CTALabel:        limit(in.CTALabel, 40, "Label tombol", &errs),
		OfferEndsAt:     in.OfferEndsAt,
	}
	if strings.TrimSpace(in.HeroImageURL) != "" {
		safe, ok := SafeImageURL(in.HeroImageURL)
		if !ok {
			errs = append(errs, "URL gambar hero harus http(s) atau path /...")
			out.HeroImageURL = strings.TrimSpace(in.HeroImageURL)
		} else {
			out.HeroImageURL = safe
		}
	}
	out.Stats = cleanStats(in.Stats, &errs)
	out.Problems = cleanLines(in.Problems, 8, 240, "Masalah", &errs)
	out.Includes = cleanLines(in.Includes, 8, 120, "Isi produk", &errs)
	out.Benefits = cleanLines(in.Benefits, 8, 200, "Manfaat", &errs)
	out.Checklist = cleanLines(in.Checklist, 8, 160, "Checklist", &errs)
	out.Bonuses = cleanBonuses(in.Bonuses, &errs)
	out.Testimonials = cleanTestimonials(in.Testimonials, &errs)
	out.FAQ = cleanFAQ(in.FAQ, &errs)
	if out.CTALabel == "" {
		out.CTALabel = "Beli sekarang"
	}
	if status == ProductPublished && out.HeroHeadline == "" {
		errs = append(errs, "Headline wajib diisi untuk produk tayang")
	}
	return out, errs
}

func cleanStats(in []Stat, errs *[]string) []Stat {
	out := make([]Stat, 0, len(in))
	for _, item := range in {
		item.Value = strings.TrimSpace(item.Value)
		item.Label = strings.TrimSpace(item.Label)
		if item.Value == "" && item.Label == "" {
			continue
		}
		if utf8.RuneCountInString(item.Value) > 24 {
			*errs = append(*errs, "Angka statistik maksimal 24 karakter")
		}
		if utf8.RuneCountInString(item.Label) > 24 {
			*errs = append(*errs, "Label statistik maksimal 24 karakter")
		}
		out = append(out, item)
	}
	if len(out) > 4 {
		*errs = append(*errs, "Statistik maksimal 4")
	}
	return out
}

func cleanBonuses(in []Bonus, errs *[]string) []Bonus {
	out := make([]Bonus, 0, len(in))
	for _, item := range in {
		item.Name = strings.TrimSpace(item.Name)
		item.Value = strings.TrimSpace(item.Value)
		if item.Name == "" && item.Value == "" {
			continue
		}
		if utf8.RuneCountInString(item.Name) > 120 {
			*errs = append(*errs, "Nama bonus maksimal 120 karakter")
		}
		if utf8.RuneCountInString(item.Value) > 80 {
			*errs = append(*errs, "Nilai bonus maksimal 80 karakter")
		}
		out = append(out, item)
	}
	if len(out) > 8 {
		*errs = append(*errs, "Bonus maksimal 8")
	}
	return out
}

func cleanTestimonials(in []Testimonial, errs *[]string) []Testimonial {
	out := make([]Testimonial, 0, len(in))
	for _, item := range in {
		item.Quote = strings.TrimSpace(item.Quote)
		item.Author = strings.TrimSpace(item.Author)
		if item.Quote == "" && item.Author == "" {
			continue
		}
		if item.Stars == 0 {
			item.Stars = 5
		}
		if item.Stars < 1 || item.Stars > 5 {
			*errs = append(*errs, "Bintang testimoni harus 1 sampai 5")
		}
		if item.Quote == "" {
			*errs = append(*errs, "Kutipan testimoni wajib diisi")
		}
		if utf8.RuneCountInString(item.Quote) > 500 {
			*errs = append(*errs, "Kutipan testimoni maksimal 500 karakter")
		}
		if utf8.RuneCountInString(item.Author) > 80 {
			*errs = append(*errs, "Nama testimoni maksimal 80 karakter")
		}
		out = append(out, item)
	}
	if len(out) > 8 {
		*errs = append(*errs, "Testimoni maksimal 8")
	}
	return out
}

func cleanFAQ(in []FAQ, errs *[]string) []FAQ {
	out := make([]FAQ, 0, len(in))
	for _, item := range in {
		item.Question = strings.TrimSpace(item.Question)
		item.Answer = strings.TrimSpace(item.Answer)
		if item.Question == "" && item.Answer == "" {
			continue
		}
		if item.Question == "" || item.Answer == "" {
			*errs = append(*errs, "Pertanyaan dan jawaban FAQ wajib diisi berpasangan")
		}
		if utf8.RuneCountInString(item.Question) > 200 {
			*errs = append(*errs, "Pertanyaan FAQ maksimal 200 karakter")
		}
		if utf8.RuneCountInString(item.Answer) > 800 {
			*errs = append(*errs, "Jawaban FAQ maksimal 800 karakter")
		}
		out = append(out, item)
	}
	if len(out) > 12 {
		*errs = append(*errs, "FAQ maksimal 12")
	}
	return out
}

func cleanLines(in []string, maxItems, maxLen int, label string, errs *[]string) []string {
	out := make([]string, 0, len(in))
	for _, raw := range in {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if utf8.RuneCountInString(s) > maxLen {
			*errs = append(*errs, label+" maksimal "+strconv.Itoa(maxLen)+" karakter")
		}
		out = append(out, s)
	}
	if len(out) > maxItems {
		*errs = append(*errs, label+" maksimal "+strconv.Itoa(maxItems)+" baris")
	}
	return out
}

func limit(s string, max int, label string, errs *[]string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > max {
		*errs = append(*errs, label+" maksimal "+strconv.Itoa(max)+" karakter")
	}
	return s
}

func unique(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
