package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidSlug(s string) bool {
	n := utf8.RuneCountInString(s)
	return n >= 2 && n <= 80 && slugPattern.MatchString(s)
}

func ReservedSlug(s string) bool {
	switch s {
	case "admin", "api", "bayar", "login", "logout", "static", "health":
		return true
	default:
		return false
	}
}

func SafeImageURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", true
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") && !strings.ContainsAny(raw, "\\ ") {
		return raw, true
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return "", false
	}
	return raw, true
}

func ValidateCustomer(name, email, phone string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < 2 || n > 100 {
		return "", "", "", errors.New("Nama harus 2–100 karakter")
	}
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil || !strings.Contains(addr.Address, "@") {
		return "", "", "", errors.New("Email tidak valid")
	}
	cleanedPhone, err := cleanPhone(phone)
	if err != nil {
		return "", "", "", err
	}
	return name, strings.ToLower(addr.Address), cleanedPhone, nil
}

func cleanPhone(s string) (string, error) {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if r == '+' && b.Len() == 0 {
			b.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	digits := strings.TrimPrefix(out, "+")
	if len(digits) < 8 || len(digits) > 15 {
		return "", errors.New("Nomor telepon tidak valid")
	}
	return out, nil
}

func FormatIDR(amount int64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	s := fmt.Sprintf("%d", amount)
	var b strings.Builder
	if neg {
		b.WriteString("-")
	}
	b.WriteString("Rp ")
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return b.String()
}

var monthsID = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// FormatJakarta renders a time the way Indonesian readers write it: "17 Okt 2026, 17.29".
func FormatJakarta(t time.Time) string {
	t = t.In(jakarta())
	return fmt.Sprintf("%d %s %d, %02d.%02d", t.Day(), monthsID[t.Month()-1], t.Year(), t.Hour(), t.Minute())
}

func FormatOfferInput(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.In(jakarta()).Format("2006-01-02T15:04")
}

func ParseOfferInput(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02T15:04", s, jakarta())
	if err != nil {
		return nil, errors.New("Waktu selesai penawaran tidak valid")
	}
	return &parsed, nil
}

func jakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Local
	}
	return loc
}

func NewInvoice(now time.Time) string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		buf = []byte{1, 2, 3, 4}
	}
	return fmt.Sprintf("PNX-%s-%s", now.In(jakarta()).Format("060102"), strings.ToUpper(hex.EncodeToString(buf)))
}

func ParseIDR(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "Rp")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, nil
	}
	var n int64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, errors.New("Harga harus berupa angka")
		}
		n = n*10 + int64(r-'0')
	}
	return n, nil
}
