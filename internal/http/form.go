package http

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"panic/internal/domain"
	"panic/internal/service"
)

type productInput struct {
	Product   domain.Product
	OfferEnds string
	Message   string
}

func (s *Server) readProductForm(c *gin.Context, actor service.Actor) productInput {
	offerRaw := c.PostForm("offer_ends_at")
	ends, offerErr := domain.ParseOfferInput(offerRaw)
	price, priceErr := domain.ParseIDR(c.PostForm("price"))
	original, originalErr := domain.ParseIDR(c.PostForm("original_price"))
	product := domain.Product{
		SellerID:      parseUint(c.PostForm("seller_id")),
		CategoryID:    parseUint(c.PostForm("category_id")),
		Name:          c.PostForm("name"),
		Slug:          strings.ToLower(strings.TrimSpace(c.PostForm("slug"))),
		Price:         price,
		OriginalPrice: original,
		Status:        c.PostForm("status"),
		ImageURL:      c.PostForm("image_url"),
		Landing: domain.Landing{
			MetaTitle:       c.PostForm("meta_title"),
			MetaDescription: c.PostForm("meta_description"),
			HeroHeadline:    c.PostForm("hero_headline"),
			HeroSubhead:     c.PostForm("hero_subhead"),
			HeroImageURL:    c.PostForm("hero_image_url"),
			Stats:           zipStats(c.PostFormArray("stat_value"), c.PostFormArray("stat_label")),
			Problems:        c.PostFormArray("problem"),
			ProblemClose:    c.PostForm("problem_close"),
			Includes:        c.PostFormArray("include"),
			IncludeNote:     c.PostForm("include_note"),
			Benefits:        c.PostFormArray("benefit"),
			Bonuses:         zipBonuses(c.PostFormArray("bonus_name"), c.PostFormArray("bonus_value")),
			Testimonials:    zipTestimonials(c.PostFormArray("testi_quote"), c.PostFormArray("testi_author"), c.PostFormArray("testi_stars")),
			UrgencyText:     c.PostForm("urgency_text"),
			OfferEndsAt:     ends,
			Checklist:       c.PostFormArray("checklist"),
			CTALabel:        c.PostForm("cta_label"),
			FAQ:             zipFAQ(c.PostFormArray("faq_q"), c.PostFormArray("faq_a")),
		},
	}
	if !actor.IsAdmin() {
		product.SellerID = actor.SellerID
	}
	var messages []string
	if offerErr != nil {
		messages = append(messages, offerErr.Error())
	}
	if priceErr != nil {
		messages = append(messages, priceErr.Error())
	}
	if originalErr != nil {
		messages = append(messages, originalErr.Error())
	}
	cleaned, err := domain.ValidateProduct(product)
	if err != nil {
		messages = append(messages, err.Error())
	}
	return productInput{Product: cleaned, OfferEnds: offerRaw, Message: strings.Join(messages, ". ")}
}

func parseUint(raw string) uint {
	n, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return uint(n)
}

func at(values []string, i int) string {
	if i < 0 || i >= len(values) {
		return ""
	}
	return values[i]
}

func zipStats(values, labels []string) []domain.Stat {
	n := len(values)
	if len(labels) > n {
		n = len(labels)
	}
	out := make([]domain.Stat, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, domain.Stat{Value: at(values, i), Label: at(labels, i)})
	}
	return out
}

func zipBonuses(names, values []string) []domain.Bonus {
	n := len(names)
	if len(values) > n {
		n = len(values)
	}
	out := make([]domain.Bonus, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, domain.Bonus{Name: at(names, i), Value: at(values, i)})
	}
	return out
}

func zipTestimonials(quotes, authors, stars []string) []domain.Testimonial {
	n := len(quotes)
	if len(authors) > n {
		n = len(authors)
	}
	if len(stars) > n {
		n = len(stars)
	}
	out := make([]domain.Testimonial, 0, n)
	for i := 0; i < n; i++ {
		star, _ := strconv.Atoi(strings.TrimSpace(at(stars, i)))
		out = append(out, domain.Testimonial{Quote: at(quotes, i), Author: at(authors, i), Stars: star})
	}
	return out
}

func zipFAQ(questions, answers []string) []domain.FAQ {
	n := len(questions)
	if len(answers) > n {
		n = len(answers)
	}
	out := make([]domain.FAQ, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, domain.FAQ{Question: at(questions, i), Answer: at(answers, i)})
	}
	return out
}

func routeID(c *gin.Context) (uint, bool) {
	id := parseUint(c.Param("id"))
	return id, id > 0
}

func formError(err error) string {
	var invalid service.ValidationError
	if errors.As(err, &invalid) {
		return invalid.Message
	}
	if errors.Is(err, service.ErrSlugTaken) {
		return "Slug sudah dipakai di toko ini."
	}
	return ""
}
