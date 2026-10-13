package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"
)

func TestVerifySignatureAmountFormats(t *testing.T) {
	orderID := "PNX-1"
	status := "200"
	server := "secret"
	sum := sha512.Sum512([]byte(orderID + status + "10000.00" + server))
	sig := hex.EncodeToString(sum[:])

	if !VerifySignature(orderID, status, "10000.00", server, sig) {
		t.Fatal("exact")
	}
	if !VerifySignature(orderID, status, "10000", server, stringsUpper(sig)) {
		t.Fatal("integer amount should match two-decimal signature")
	}
	if VerifySignature(orderID, status, "10000.00", "other", sig) {
		t.Fatal("wrong key")
	}
	if VerifySignature(orderID, status, "10000.00", server, "") {
		t.Fatal("empty signature")
	}
}

func stringsUpper(s string) string {
	return hexUpper(s)
}

func hexUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'f' {
			b[i] = c - 32
		}
	}
	return string(b)
}
