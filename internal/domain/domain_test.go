package domain

import (
	"strings"
	"testing"
)

func TestValidateProductPublishedRequiresHeadlineAndPrice(t *testing.T) {
	_, err := ValidateProduct(Product{
		SellerID:   1,
		CategoryID: 1,
		Name:       "Paket",
		Slug:       "paket",
		Status:     ProductPublished,
		Price:      0,
	})
	if err == nil || !strings.Contains(err.Error(), "Harga") || !strings.Contains(err.Error(), "Headline") {
		t.Fatalf("got %v", err)
	}
}

func TestCleanLandingDropsEmptyAndDefaultsCTA(t *testing.T) {
	landing, errs := CleanLanding(Landing{
		HeroHeadline: "Judul",
		Problems:     []string{"  Satu  ", ""},
		Stats:        []Stat{{}, {Value: "48", Label: "Video"}},
		CTALabel:     "",
	}, ProductDraft)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	if landing.CTALabel != "Beli sekarang" {
		t.Fatalf("cta %q", landing.CTALabel)
	}
	if len(landing.Problems) != 1 || landing.Problems[0] != "Satu" {
		t.Fatalf("problems %#v", landing.Problems)
	}
	if len(landing.Stats) != 1 {
		t.Fatalf("stats %#v", landing.Stats)
	}
}

func TestCleanLandingRejectsTooManyStats(t *testing.T) {
	_, errs := CleanLanding(Landing{
		Stats: []Stat{
			{Value: "1", Label: "A"},
			{Value: "2", Label: "B"},
			{Value: "3", Label: "C"},
			{Value: "4", Label: "D"},
			{Value: "5", Label: "E"},
		},
	}, ProductDraft)
	if len(errs) == 0 {
		t.Fatal("expected max stats error")
	}
}

func TestFormatIDR(t *testing.T) {
	if got := FormatIDR(99000); got != "Rp 99.000" {
		t.Fatal(got)
	}
	if got := FormatIDR(1000000); got != "Rp 1.000.000" {
		t.Fatal(got)
	}
}

func TestParseIDR(t *testing.T) {
	n, err := ParseIDR("Rp 99.000")
	if err != nil || n != 99000 {
		t.Fatalf("%d %v", n, err)
	}
}

func TestReservedSlug(t *testing.T) {
	if !ReservedSlug("admin") || ReservedSlug("toko-contoh") {
		t.Fatal("reserved mismatch")
	}
	if !ValidSlug("paket-belajar") || ValidSlug("Paket") || ValidSlug("-paket") {
		t.Fatal("slug mismatch")
	}
}

func TestNextStatusDoesNotDowngradePaid(t *testing.T) {
	if next, ok := NextStatus(StatusSettlement, StatusPending); ok || next != StatusSettlement {
		t.Fatalf("next %s ok %v", next, ok)
	}
	if next, ok := NextStatus(StatusPending, StatusCapture); !ok || next != StatusCapture {
		t.Fatalf("next %s ok %v", next, ok)
	}
}

func TestNormalizeChallengeCaptureStaysPending(t *testing.T) {
	if got := NormalizePaymentStatus("capture", "challenge"); got != StatusPending {
		t.Fatal(got)
	}
	if got := NormalizePaymentStatus("capture", "accept"); got != StatusCapture {
		t.Fatal(got)
	}
}

func TestValidateCustomer(t *testing.T) {
	name, email, phone, err := ValidateCustomer(" Rani ", "Rani@Example.com", "+62 812-0000-1111")
	if err != nil {
		t.Fatal(err)
	}
	if name != "Rani" || email != "rani@example.com" || phone != "+6281200001111" {
		t.Fatalf("%s %s %s", name, email, phone)
	}
}

func TestParseOfferInput(t *testing.T) {
	got, err := ParseOfferInput("2026-10-20T21:30")
	if err != nil || got == nil {
		t.Fatal(err)
	}
	if got.In(jakarta()).Format("15:04") != "21:30" {
		t.Fatal(got)
	}
	empty, err := ParseOfferInput("  ")
	if err != nil || empty != nil {
		t.Fatal(empty, err)
	}
}
