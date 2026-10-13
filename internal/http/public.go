package http

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"panic/internal/domain"
	"panic/internal/service"
)

func (s *Server) home(c *gin.Context) {
	stores, err := s.catalog.PublicStores()
	if err != nil {
		s.fail(c, err)
		return
	}
	page := homeView{view: s.baseView(c, "Toko produk digital")}
	page.Description = "Daftar toko produk digital yang sedang tayang."
	if s.cfg.AppURL != "" {
		page.Canonical = s.cfg.AppURL + "/"
	}
	page.Stores = make([]storeCard, 0, len(stores))
	for _, store := range stores {
		page.Stores = append(page.Stores, storeCard{
			Name:        store.StoreName,
			Slug:        store.Slug,
			Description: store.Description,
		})
	}
	s.render(c, "page-home", page)
}

func (s *Server) store(c *gin.Context) {
	seller, products, err := s.catalog.Store(c.Param("seller"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			s.notFound(c)
			return
		}
		s.fail(c, err)
		return
	}
	page := storeView{
		view:        s.baseView(c, seller.StoreName),
		StoreName:   seller.StoreName,
		Description: seller.Description,
	}
	page.Description = seller.Description
	if s.cfg.AppURL != "" {
		page.Canonical = s.cfg.AppURL + "/" + seller.Slug
	}
	page.Products = make([]productCard, 0, len(products))
	for _, product := range products {
		image := product.ImageURL
		if image == "" {
			image = product.Landing.HeroImageURL
		}
		page.Products = append(page.Products, productCard{
			Name:          product.Name,
			Href:          "/" + seller.Slug + "/" + product.Slug,
			Summary:       summaryOf(product),
			Image:         image,
			Price:         product.Price,
			OriginalPrice: product.OriginalPrice,
		})
	}
	s.render(c, "page-store", page)
}

func (s *Server) landing(c *gin.Context) {
	seller, product, err := s.catalog.Published(c.Param("seller"), c.Param("product"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			s.notFound(c)
			return
		}
		s.fail(c, err)
		return
	}
	s.render(c, "page-landing", s.landingPage(c, seller, product))
}

func (s *Server) landingPage(c *gin.Context, seller domain.Seller, product domain.Product) landingView {
	landing := product.Landing
	headline := landing.HeroHeadline
	if headline == "" {
		headline = product.Name
	}
	description := landing.MetaDescription
	if description == "" {
		description = landing.HeroSubhead
	}
	title := landing.MetaTitle
	if title == "" {
		title = headline
	}
	heroImage := landing.HeroImageURL
	if heroImage == "" {
		heroImage = product.ImageURL
	}
	path := "/" + seller.Slug + "/" + product.Slug
	base := s.publicBase(c)
	ogImage := absolute(base, heroImage)
	quotes := make([]testiView, 0, len(landing.Testimonials))
	for _, item := range landing.Testimonials {
		quotes = append(quotes, testiView{Quote: item.Quote, Author: item.Author, Stars: item.Stars})
	}
	page := landingView{
		view:          s.baseView(c, title),
		StoreName:     seller.StoreName,
		StoreSlug:     seller.Slug,
		CheckoutURL:   path + "/checkout",
		HeroHeadline:  headline,
		HeroSubhead:   landing.HeroSubhead,
		HeroImage:     heroImage,
		Stats:         landing.Stats,
		Problems:      landing.Problems,
		ProblemClose:  landing.ProblemClose,
		Includes:      landing.Includes,
		IncludeNote:   landing.IncludeNote,
		Benefits:      landing.Benefits,
		Bonuses:       landing.Bonuses,
		Testimonials:  quotes,
		Urgency:       landing.UrgencyText,
		Checklist:     landing.Checklist,
		CTA:           landing.CTALabel,
		Price:         product.Price,
		OriginalPrice: product.OriginalPrice,
		FAQ:           landing.FAQ,
	}
	if page.CTA == "" {
		page.CTA = "Beli sekarang"
	}
	page.Description = description
	page.Image = ogImage
	page.BodyClass = "has-buybar"
	page.JSONLD = productJSONLD(product, description, ogImage)
	if s.cfg.AppURL != "" {
		page.Canonical = s.cfg.AppURL + path
	}
	if landing.OfferEndsAt != nil && !landing.OfferEndsAt.IsZero() {
		page.HasCountdown = true
		page.OfferEnds = landing.OfferEndsAt.Format(time.RFC3339)
	}
	return page
}

func (s *Server) fail(c *gin.Context, err error) {
	if err == nil {
		return
	}
	log.Printf("request failed: %v", err)
	s.renderStatus(c, http.StatusInternalServerError, "page-message", messageView{
		view:    s.baseView(c, "Terjadi kesalahan"),
		Message: "Halaman gagal dimuat. Coba lagi beberapa saat.",
	})
}
