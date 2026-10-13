package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"panic/internal/domain"
	"panic/internal/midtrans"
)

type Payments struct {
	db        *gorm.DB
	client    *midtrans.Client
	serverKey string
}

func NewPayments(db *gorm.DB, serverKey, clientKey string, production bool) *Payments {
	return &Payments{
		db:        db,
		client:    midtrans.NewClient(serverKey, clientKey, production),
		serverKey: serverKey,
	}
}

func (s *Payments) Enabled() bool {
	return s.client.Configured()
}

func (s *Payments) ClientKey() string {
	return s.client.ClientKey()
}

func (s *Payments) SnapJS() string {
	return s.client.SnapJSURL()
}

func (s *Payments) EnsureSnap(order *domain.Order, finishURL string) error {
	if !s.Enabled() {
		return ErrNotConfigured
	}
	if order == nil || !domain.CanPay(order.Status) || order.SnapToken != "" {
		return nil
	}
	orderID := order.MidtransOrderID
	if orderID == "" {
		orderID = order.Invoice
	}
	result, err := s.client.CreateSnap(midtrans.SnapInput{
		OrderID:     orderID,
		GrossAmount: order.GrossAmount,
		Name:        order.CustomerName,
		Email:       order.CustomerEmail,
		Phone:       order.CustomerPhone,
		ItemID:      strconv.FormatUint(uint64(order.ProductID), 10),
		ItemName:    order.ProductName,
		FinishURL:   finishURL,
	})
	if err != nil {
		return err
	}
	if err := s.db.Model(&domain.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"snap_token":        result.Token,
		"snap_redirect_url": result.RedirectURL,
		"midtrans_order_id": orderID,
	}).Error; err != nil {
		return err
	}
	order.SnapToken = result.Token
	order.SnapRedirectURL = result.RedirectURL
	order.MidtransOrderID = orderID
	return nil
}

func (s *Payments) Sync(order *domain.Order) error {
	if !s.Enabled() || order == nil || !domain.CanPay(order.Status) {
		return nil
	}
	orderID := order.MidtransOrderID
	if orderID == "" {
		orderID = order.Invoice
	}
	note, err := s.client.Status(orderID)
	if err != nil {
		return err
	}
	return s.applyNote(order, note)
}

func (s *Payments) HandleNotification(body []byte) error {
	if !s.Enabled() {
		return ErrNotConfigured
	}
	var note midtrans.Notification
	if err := json.Unmarshal(body, &note); err != nil {
		return err
	}
	return s.applyIncoming(note)
}

func (s *Payments) applyNote(order *domain.Order, note midtrans.Notification) error {
	if !midtrans.VerifySignature(note.OrderID, note.StatusCode, note.GrossAmount, s.serverKey, note.SignatureKey) {
		return ErrBadSignature
	}
	if !amountsMatch(order.GrossAmount, note.GrossAmount) {
		return ErrAmountMismatch
	}
	incoming := domain.NormalizePaymentStatus(note.TransactionStatus, note.FraudStatus)
	if incoming == "" {
		return nil
	}
	return s.apply(order, incoming)
}

func (s *Payments) applyIncoming(note midtrans.Notification) error {
	if !midtrans.VerifySignature(note.OrderID, note.StatusCode, note.GrossAmount, s.serverKey, note.SignatureKey) {
		return ErrBadSignature
	}
	var order domain.Order
	err := s.db.Where("midtrans_order_id = ? OR invoice = ?", note.OrderID, note.OrderID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return s.applyNote(&order, note)
}

func (s *Payments) apply(order *domain.Order, incoming string) error {
	next, ok := domain.NextStatus(order.Status, incoming)
	if !ok {
		return nil
	}
	updates := map[string]any{"status": next}
	var paidAt *time.Time
	if domain.IsPaidStatus(next) && order.PaidAt == nil {
		now := time.Now()
		paidAt = &now
		updates["paid_at"] = now
	}
	result := s.db.Model(&domain.Order{}).Where("id = ? AND status = ?", order.ID, order.Status).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var fresh domain.Order
		if err := s.db.First(&fresh, order.ID).Error; err != nil {
			return err
		}
		*order = fresh
		return nil
	}
	order.Status = next
	if paidAt != nil {
		order.PaidAt = paidAt
	}
	return nil
}

func amountsMatch(expected int64, gross string) bool {
	gross = strings.TrimSpace(gross)
	if gross == "" {
		return false
	}
	value, err := strconv.ParseFloat(gross, 64)
	if err != nil {
		return false
	}
	return int64(value) == expected
}
