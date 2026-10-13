package http

import (
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"panic/internal/domain"
	"panic/internal/service"
)

var invoicePattern = regexp.MustCompile(`^PNX-[0-9]{6}-[A-F0-9]{8}$`)

func (s *Server) checkoutForm(c *gin.Context) {
	seller, product, err := s.catalog.Published(c.Param("seller"), c.Param("product"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			s.notFound(c)
			return
		}
		s.fail(c, err)
		return
	}
	s.render(c, "page-checkout", s.checkoutPage(c, seller, product, "", "", "", ""))
}

func (s *Server) checkoutSubmit(c *gin.Context) {
	seller, product, err := s.catalog.Published(c.Param("seller"), c.Param("product"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			s.notFound(c)
			return
		}
		s.fail(c, err)
		return
	}
	name := c.PostForm("name")
	email := c.PostForm("email")
	phone := c.PostForm("phone")
	order, err := s.orders.Create(product, name, email, phone)
	if err != nil {
		var invalid service.ValidationError
		if errors.As(err, &invalid) {
			s.renderStatus(c, http.StatusUnprocessableEntity, "page-checkout", s.checkoutPage(c, seller, product, name, email, phone, invalid.Message))
			return
		}
		s.fail(c, err)
		return
	}
	finish := s.publicBase(c) + "/bayar/" + order.Invoice
	if err := s.payments.EnsureSnap(&order, finish); err != nil && !errors.Is(err, service.ErrNotConfigured) {
		log.Printf("snap %s: %v", order.Invoice, err)
	}
	c.Redirect(http.StatusSeeOther, "/bayar/"+order.Invoice)
}

func (s *Server) checkoutPage(c *gin.Context, seller domain.Seller, product domain.Product, name, email, phone, message string) checkoutView {
	page := checkoutView{
		view:        s.baseView(c, "Checkout — "+product.Name),
		StoreSlug:   seller.Slug,
		ProductSlug: product.Slug,
		ProductName: product.Name,
		Price:       product.Price,
		Name:        name,
		Email:       email,
		Phone:       phone,
		Error:       message,
	}
	page.Robots = "noindex, nofollow"
	page.Description = "Isi nama, email, dan telepon untuk melanjutkan pembayaran."
	return page
}

func (s *Server) payment(c *gin.Context) {
	invoice := c.Param("invoice")
	if !invoicePattern.MatchString(invoice) {
		s.notFound(c)
		return
	}
	order, err := s.orders.ByInvoice(invoice)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			s.notFound(c)
			return
		}
		s.fail(c, err)
		return
	}
	if order.SnapToken != "" {
		if err := s.payments.Sync(&order); err != nil {
			log.Printf("sync %s: %v", order.Invoice, err)
		}
	}
	note := ""
	finish := s.publicBase(c) + "/bayar/" + order.Invoice
	if err := s.payments.EnsureSnap(&order, finish); err != nil {
		if errors.Is(err, service.ErrNotConfigured) {
			note = "Midtrans belum diatur. Pesanan tetap tersimpan."
		} else if domain.CanPay(order.Status) {
			log.Printf("snap %s: %v", order.Invoice, err)
			note = "Token pembayaran gagal dibuat. Muat ulang halaman ini beberapa saat lagi."
		}
	}
	page := paymentView{
		view:        s.baseView(c, "Status pembayaran"),
		Invoice:     order.Invoice,
		ProductName: order.ProductName,
		Email:       order.CustomerEmail,
		Amount:      order.GrossAmount,
		Status:      order.Status,
		StatusLabel: domain.StatusLabel(order.Status),
		Note:        note,
		Pending:     domain.CanPay(order.Status),
		RedirectURL: safeSnapURL(order.SnapRedirectURL),
	}
	page.Robots = "noindex, nofollow"
	if page.Pending && order.SnapToken != "" && s.payments.ClientKey() != "" {
		page.HasSnap = true
		page.SnapJS = s.payments.SnapJS()
		page.ClientKey = s.payments.ClientKey()
		page.Token = order.SnapToken
	}
	s.render(c, "page-payment", page)
}

func (s *Server) notification(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	err = s.payments.HandleNotification(body)
	switch {
	case err == nil:
		c.Status(http.StatusOK)
	case errors.Is(err, service.ErrBadSignature), errors.Is(err, service.ErrAmountMismatch):
		c.Status(http.StatusForbidden)
	case errors.Is(err, service.ErrNotFound):
		c.Status(http.StatusNotFound)
	case errors.Is(err, service.ErrNotConfigured):
		c.Status(http.StatusServiceUnavailable)
	default:
		log.Printf("midtrans notification: %v", err)
		c.Status(http.StatusInternalServerError)
	}
}
