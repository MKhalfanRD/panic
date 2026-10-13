package domain

import "strings"

const (
	StatusPending    = "pending"
	StatusSettlement = "settlement"
	StatusCapture    = "capture"
	StatusPaid       = "paid"
	StatusExpire     = "expire"
	StatusCancel     = "cancel"
	StatusDeny       = "deny"
)

func IsPaidStatus(status string) bool {
	switch status {
	case StatusPaid, StatusSettlement, StatusCapture:
		return true
	default:
		return false
	}
}

func CanPay(status string) bool {
	return status == StatusPending
}

func StatusLabel(status string) string {
	switch status {
	case StatusPending:
		return "Menunggu pembayaran"
	case StatusPaid, StatusSettlement, StatusCapture:
		return "Lunas"
	case StatusExpire:
		return "Kedaluwarsa"
	case StatusCancel:
		return "Dibatalkan"
	case StatusDeny:
		return "Ditolak"
	default:
		return status
	}
}

func NormalizePaymentStatus(transactionStatus, fraudStatus string) string {
	ts := strings.ToLower(strings.TrimSpace(transactionStatus))
	fraud := strings.ToLower(strings.TrimSpace(fraudStatus))
	if fraud == "deny" {
		return StatusDeny
	}
	switch ts {
	case StatusCapture:
		if fraud == "challenge" {
			return StatusPending
		}
		return StatusCapture
	case StatusSettlement:
		return StatusSettlement
	case StatusPaid:
		return StatusPaid
	case StatusPending:
		return StatusPending
	case StatusExpire:
		return StatusExpire
	case StatusCancel, "refund", "partial_refund":
		return StatusCancel
	case StatusDeny:
		return StatusDeny
	default:
		return ""
	}
}

func NextStatus(current, incoming string) (string, bool) {
	switch incoming {
	case StatusPending, StatusSettlement, StatusCapture, StatusPaid, StatusExpire, StatusCancel, StatusDeny:
	default:
		return current, false
	}
	if IsPaidStatus(current) && !IsPaidStatus(incoming) {
		return current, false
	}
	if current == incoming {
		return current, false
	}
	return incoming, true
}
