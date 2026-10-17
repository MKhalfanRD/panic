package http

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"panic/internal/domain"
	"panic/web"
)

type view struct {
	Title        string
	Description  string
	Canonical    string
	Image        string
	Robots       string
	BodyClass    string
	JSONLD       template.JS
	CSRF         string
	UserName     string
	IsAdmin      bool
	DevHint      bool
	HasCountdown bool
}

type homeView struct {
	view
	Offers   []productCard
	Featured *productCard
	Stores   []storeCard
	MinPrice int64
}

type storeCard struct {
	Name        string
	Href        string
	Description string
	Count       int
}

type storeView struct {
	view
	StoreName   string
	Description string
	Products    []productCard
}

type productCard struct {
	StoreName     string
	Name          string
	Href          string
	Summary       string
	Image         string
	Mark          string
	Price         int64
	OriginalPrice int64
}

type landingView struct {
	view
	StoreName     string
	StoreSlug     string
	CheckoutURL   string
	HeroHeadline  string
	HeroSubhead   string
	HeroImage     string
	Stats         []domain.Stat
	Problems      []string
	ProblemClose  string
	Includes      []string
	IncludeNote   string
	Benefits      []string
	Bonuses       []domain.Bonus
	Testimonials  []testiView
	Urgency       string
	OfferEnds     string
	Checklist     []string
	CTA           string
	Price         int64
	OriginalPrice int64
	FAQ           []domain.FAQ
}

type testiView struct {
	Quote  string
	Author string
	Stars  int
}

type checkoutView struct {
	view
	StoreSlug     string
	StoreName     string
	ProductSlug   string
	ProductName   string
	Price         int64
	OriginalPrice int64
	Name          string
	Email       string
	Phone       string
	Error       string
}

type paymentView struct {
	view
	Invoice     string
	ProductName string
	StoreName   string
	StoreSlug   string
	CreatedAt   string
	PaidAt      string
	Email       string
	Amount      int64
	Status      string
	StatusLabel string
	Note        string
	HasSnap     bool
	SnapJS      string
	ClientKey   string
	Token       string
	RedirectURL string
	Pending     bool
}

type loginView struct {
	view
	Error string
	Next  string
}

type messageView struct {
	view
	Message string
}

type productListView struct {
	view
	Products []domain.Product
}

type productFormView struct {
	view
	Product    domain.Product
	OfferEnds  string
	Categories []domain.Category
	Sellers    []domain.Seller
	Error      string
	Saved      bool
	IsNew      bool
	PublicPath string
}

type orderListView struct {
	view
	Orders []domain.Order
}

func ParseTemplates() (*template.Template, error) {
	return template.New("root").Funcs(template.FuncMap{
		"formatIDR":   domain.FormatIDR,
		"formatTime":  domain.FormatJakarta,
		"statusLabel": domain.StatusLabel,
		"stars":       stars,
		"sub":         func(a, b int64) int64 { return a - b },
		"statusKind":  statusKind,
		"initial":     initial,
	}).ParseFS(web.Files, "templates/*.html")
}

// statusKind groups payment statuses into the three looks the UI has.
func statusKind(status string) string {
	switch {
	case domain.IsPaidStatus(status):
		return "paid"
	case domain.CanPay(status):
		return "pending"
	default:
		return "closed"
	}
}

func initial(name string) string {
	for _, r := range strings.TrimSpace(name) {
		return strings.ToUpper(string(r))
	}
	return ""
}

func stars(n int) []bool {
	if n < 0 {
		n = 0
	}
	if n > 5 {
		n = 5
	}
	out := make([]bool, 5)
	for i := 0; i < n; i++ {
		out[i] = true
	}
	return out
}

func (s *Server) render(c *gin.Context, name string, data any) {
	s.renderStatus(c, http.StatusOK, name, data)
}

func (s *Server) renderStatus(c *gin.Context, code int, name string, data any) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("template %s: %v", name, err)
		c.String(http.StatusInternalServerError, "Gagal menampilkan halaman")
		return
	}
	c.Data(code, "text/html; charset=utf-8", buf.Bytes())
}

func (s *Server) baseView(c *gin.Context, title string) view {
	v := view{
		Title:  title,
		Robots: "index, follow",
		CSRF:   s.csrfToken(c),
	}
	if actor, ok := s.currentActor(c); ok {
		v.UserName = actor.User.Name
		v.IsAdmin = actor.IsAdmin()
	}
	return v
}

func (s *Server) adminView(c *gin.Context, title string) view {
	v := s.baseView(c, title)
	v.Robots = "noindex, nofollow"
	v.BodyClass = "admin"
	return v
}

func (s *Server) notFound(c *gin.Context) {
	s.renderStatus(c, http.StatusNotFound, "page-message", messageView{
		view:    s.baseView(c, "Halaman tidak ditemukan"),
		Message: "Link mungkin salah ketik, atau produknya sudah tidak dijual. Untuk status pesanan, buka link halaman pembayaran yang Anda simpan.",
	})
}

func productJSONLD(product domain.Product, description, image string) template.JS {
	payload := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "Product",
		"name":        product.Name,
		"description": description,
		"offers": map[string]any{
			"@type":         "Offer",
			"priceCurrency": "IDR",
			"price":         product.Price,
			"availability":  "https://schema.org/InStock",
		},
	}
	if image != "" {
		payload["image"] = image
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return template.JS(raw)
}

func absolute(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	if strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") {
		return ref
	}
	if strings.HasPrefix(ref, "/") && base != "" {
		return strings.TrimRight(base, "/") + ref
	}
	return ""
}

func safeNext(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "/admin") || strings.HasPrefix(raw, "//") || strings.Contains(raw, "..") || strings.ContainsAny(raw, "\\?#") {
		return "/admin/products"
	}
	return raw
}

func safeSnapURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "https://app.midtrans.com/") || strings.HasPrefix(raw, "https://app.sandbox.midtrans.com/") {
		return raw
	}
	return ""
}

func offerCard(product domain.Product, href string) productCard {
	image := product.ImageURL
	if image == "" {
		image = product.Landing.HeroImageURL
	}
	return productCard{
		Name:          product.Name,
		Href:          href,
		Summary:       summaryOf(product),
		Image:         image,
		Mark:          markOf(product.Name),
		Price:         product.Price,
		OriginalPrice: product.OriginalPrice,
	}
}

func markOf(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "P"
	}
	r, _ := utf8.DecodeRuneInString(name)
	return strings.ToUpper(string(r))
}

func summaryOf(product domain.Product) string {
	for _, text := range []string{product.Landing.HeroSubhead, product.Landing.MetaDescription, product.Landing.ProblemClose} {
		if strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}
