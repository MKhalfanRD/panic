package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"panic/internal/domain"
)

type Orders struct {
	db *gorm.DB
}

func NewOrders(db *gorm.DB) *Orders {
	return &Orders{db: db}
}

func (s *Orders) Create(product domain.Product, name, email, phone string) (domain.Order, error) {
	if product.Status != domain.ProductPublished || product.Price <= 0 {
		return domain.Order{}, ErrNotFound
	}
	name, email, phone, err := domain.ValidateCustomer(name, email, phone)
	if err != nil {
		return domain.Order{}, ValidationError{Message: err.Error()}
	}
	order := domain.Order{
		SellerID:      product.SellerID,
		ProductID:     product.ID,
		ProductName:   product.Name,
		CustomerName:  name,
		CustomerEmail: email,
		CustomerPhone: phone,
		GrossAmount:   product.Price,
		Status:        domain.StatusPending,
	}
	for attempt := 0; attempt < 3; attempt++ {
		order.Invoice = domain.NewInvoice(time.Now())
		order.MidtransOrderID = order.Invoice
		err = s.db.Omit("Seller").Create(&order).Error
		if err == nil {
			return order, nil
		}
		if !isDuplicate(err) {
			return domain.Order{}, err
		}
	}
	return domain.Order{}, fmt.Errorf("gagal membuat nomor pesanan")
}

func (s *Orders) ByInvoice(invoice string) (domain.Order, error) {
	var order domain.Order
	err := s.db.Preload("Seller").Where("invoice = ?", invoice).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Order{}, ErrNotFound
	}
	return order, err
}

func (s *Orders) List(actor Actor) ([]domain.Order, error) {
	q := s.db.Preload("Seller").Order("created_at desc").Limit(200)
	if !actor.IsAdmin() {
		q = q.Where("seller_id = ?", actor.SellerID)
	}
	var orders []domain.Order
	err := q.Find(&orders).Error
	return orders, err
}
