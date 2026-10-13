package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"panic/internal/domain"
	"panic/internal/service"
)

func (s *Server) productList(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	products, err := s.catalog.List(actor)
	if err != nil {
		s.fail(c, err)
		return
	}
	page := productListView{view: s.adminView(c, "Produk"), Products: products}
	s.render(c, "page-products", page)
}

func (s *Server) productNew(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	product := domain.Product{Status: domain.ProductDraft, SellerID: actor.SellerID}
	categories, _ := s.catalog.Categories()
	if len(categories) == 1 {
		product.CategoryID = categories[0].ID
	}
	page := productFormView{
		view:       s.adminView(c, "Produk baru"),
		Product:    product,
		Categories: categories,
		IsNew:      true,
	}
	if actor.IsAdmin() {
		page.Sellers, _ = s.catalog.Sellers()
	}
	s.render(c, "page-product-form", page)
}

func (s *Server) productCreate(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	input := s.readProductForm(c, actor)
	if input.Message != "" {
		s.renderProductForm(c, actor, input.Product, input.OfferEnds, input.Message, true, "")
		return
	}
	if err := s.catalog.Save(actor, &input.Product); err != nil {
		if msg := formError(err); msg != "" {
			s.renderProductForm(c, actor, input.Product, input.OfferEnds, msg, true, "")
			return
		}
		s.fail(c, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/products/"+itoa(input.Product.ID)+"/edit?saved=1")
}

func (s *Server) productEdit(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	id, valid := routeID(c)
	if !valid {
		s.notFound(c)
		return
	}
	product, err := s.catalog.Get(actor, id)
	if err != nil {
		s.productErr(c, err)
		return
	}
	path := ""
	if product.Status == domain.ProductPublished && product.Seller.Slug != "" {
		path = "/" + product.Seller.Slug + "/" + product.Slug
	}
	s.renderProductForm(c, actor, product, domain.FormatOfferInput(product.Landing.OfferEndsAt), "", false, path)
}

func (s *Server) productUpdate(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	id, valid := routeID(c)
	if !valid {
		s.notFound(c)
		return
	}
	if _, err := s.catalog.Get(actor, id); err != nil {
		s.productErr(c, err)
		return
	}
	input := s.readProductForm(c, actor)
	input.Product.ID = id
	if input.Message != "" {
		s.renderProductForm(c, actor, input.Product, input.OfferEnds, input.Message, false, "")
		return
	}
	if err := s.catalog.Save(actor, &input.Product); err != nil {
		if msg := formError(err); msg != "" {
			s.renderProductForm(c, actor, input.Product, input.OfferEnds, msg, false, "")
			return
		}
		s.fail(c, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/products/"+itoa(id)+"/edit?saved=1")
}

func (s *Server) productDelete(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	id, valid := routeID(c)
	if !valid {
		s.notFound(c)
		return
	}
	if err := s.catalog.Delete(actor, id); err != nil {
		s.productErr(c, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/products")
}

func (s *Server) orderList(c *gin.Context) {
	actor, ok := s.mustActor(c)
	if !ok {
		return
	}
	orders, err := s.orders.List(actor)
	if err != nil {
		s.fail(c, err)
		return
	}
	s.render(c, "page-orders", orderListView{view: s.adminView(c, "Pesanan"), Orders: orders})
}

func (s *Server) renderProductForm(c *gin.Context, actor service.Actor, product domain.Product, offerEnds, message string, isNew bool, publicPath string) {
	page := productFormView{
		view:       s.adminView(c, "Produk"),
		Product:    product,
		OfferEnds:  offerEnds,
		Error:      message,
		IsNew:      isNew,
		PublicPath: publicPath,
		Saved:      message == "" && c.Query("saved") == "1",
	}
	if isNew {
		page.Title = "Produk baru"
	} else {
		page.Title = "Ubah produk"
	}
	page.Categories, _ = s.catalog.Categories()
	if actor.IsAdmin() {
		page.Sellers, _ = s.catalog.Sellers()
	}
	code := http.StatusOK
	if message != "" {
		code = http.StatusUnprocessableEntity
	}
	s.renderStatus(c, code, "page-product-form", page)
}

func (s *Server) productErr(c *gin.Context, err error) {
	if errors.Is(err, service.ErrNotFound) {
		s.notFound(c)
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		s.renderStatus(c, http.StatusForbidden, "page-message", messageView{
			view:    s.adminView(c, "Tidak diizinkan"),
			Message: "Anda tidak dapat mengubah produk ini.",
		})
		return
	}
	s.fail(c, err)
}

func itoa(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}
