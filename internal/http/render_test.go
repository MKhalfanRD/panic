package http

import (
	"bytes"
	"testing"
	"time"

	"panic/internal/domain"
)

func TestTemplatesExecute(t *testing.T) {
	tmpl, err := ParseTemplates()
	if err != nil {
		t.Fatal(err)
	}
	ends := time.Now().Add(time.Hour)
	pages := map[string]any{
		"page-home": homeView{view: view{Title: "Beranda"}},
		"page-store": storeView{view: view{Title: "Toko"}, Products: []productCard{{
			Name: "Paket", Href: "/toko/paket", Price: 1000, OriginalPrice: 2000,
		}}},
		"page-landing": landingView{
			view: view{Title: "LP", HasCountdown: true}, CheckoutURL: "/toko/paket/checkout",
			HeroHeadline: "Halo", CTA: "Beli", Price: 99000, OriginalPrice: 149000,
			OfferEnds: ends.Format(time.RFC3339),
			Stats:     []domain.Stat{{Value: "1", Label: "Video"}},
			Problems:  []string{"Satu"}, Includes: []string{"Isi"}, Benefits: []string{"Manfaat"},
			Bonuses:      []domain.Bonus{{Name: "Bonus", Value: "Rp 1"}},
			Testimonials: []testiView{{Quote: "Bagus", Author: "Rani", Stars: 5}},
			Checklist:    []string{"Satu"}, FAQ: []domain.FAQ{{Question: "Q", Answer: "A"}},
		},
		"page-checkout": checkoutView{view: view{Title: "Checkout", CSRF: "x"}, ProductName: "Paket", Price: 1000},
		"page-payment":  paymentView{view: view{Title: "Bayar"}, Invoice: "PNX-1", Amount: 1000, Status: "pending", StatusLabel: "Menunggu", Pending: true},
		"page-login":    loginView{view: view{Title: "Masuk", DevHint: true, CSRF: "x"}, Next: "/admin/products"},
		"page-message":  messageView{view: view{Title: "Tidak ada"}, Message: "Kosong"},
		"page-products": productListView{view: view{Title: "Produk", IsAdmin: true, CSRF: "x"}, Products: []domain.Product{{
			Name: "Paket", Price: 1000, Status: domain.ProductPublished, Seller: domain.Seller{StoreName: "Toko"},
		}}},
		"page-orders": orderListView{view: view{Title: "Pesanan", IsAdmin: true, CSRF: "x"}, Orders: []domain.Order{{
			Invoice: "PNX-1", ProductName: "Paket", CustomerName: "Rani", GrossAmount: 1000, Status: domain.StatusPending, Seller: domain.Seller{StoreName: "Toko"}, CreatedAt: time.Now(),
		}}},
		"page-product-form": productFormView{
			view: view{Title: "Form", CSRF: "x", IsAdmin: true},
			Product: domain.Product{
				ID: 1, Name: "Paket", Slug: "paket", Status: domain.ProductDraft, Price: 1000,
				Landing: domain.Landing{
					HeroHeadline: "Halo",
					Stats:        []domain.Stat{{Value: "1", Label: "A"}},
					Problems:     []string{"P"},
					Testimonials: []domain.Testimonial{{Quote: "Q", Author: "A", Stars: 5}},
					FAQ:          []domain.FAQ{{Question: "Q", Answer: "A"}},
				},
			},
			Categories: []domain.Category{{ID: 1, Name: "Pendidikan"}},
			Sellers:    []domain.Seller{{ID: 1, StoreName: "Toko"}},
			IsNew:      false,
		},
	}
	for name, data := range pages {
		if tmpl.Lookup(name) == nil {
			t.Errorf("missing %s", name)
			continue
		}
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if buf.Len() == 0 {
			t.Errorf("%s empty", name)
		}
	}
}
