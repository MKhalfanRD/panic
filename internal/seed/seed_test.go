package seed

import (
	"testing"
	"time"

	"panic/internal/domain"
)

func TestSampleProductIsValid(t *testing.T) {
	ends := time.Now().Add(24 * time.Hour)
	product := domain.Product{
		SellerID:      1,
		CategoryID:    1,
		Name:          "Paket Belajar Riang",
		Slug:          productSlug,
		Price:         99000,
		OriginalPrice: 149000,
		Status:        domain.ProductPublished,
		Landing:       sampleLanding(&ends),
	}
	if _, err := domain.ValidateProduct(product); err != nil {
		t.Fatal(err)
	}
}
