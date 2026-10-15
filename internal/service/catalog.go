package service

import (
	"errors"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"panic/internal/domain"
)

type Catalog struct {
	db *gorm.DB
}

func NewCatalog(db *gorm.DB) *Catalog {
	return &Catalog{db: db}
}

func (s *Catalog) PublicOffers() ([]domain.Product, error) {
	var products []domain.Product
	err := s.db.Preload("Seller").Where("status = ?", domain.ProductPublished).Order("name asc").Find(&products).Error
	return products, err
}

func (s *Catalog) Store(slug string) (domain.Seller, []domain.Product, error) {
	if domain.ReservedSlug(slug) || !domain.ValidSlug(slug) {
		return domain.Seller{}, nil, ErrNotFound
	}
	var seller domain.Seller
	if err := s.db.Where("slug = ?", slug).First(&seller).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Seller{}, nil, ErrNotFound
		}
		return domain.Seller{}, nil, err
	}
	var products []domain.Product
	err := s.db.Where("seller_id = ? AND status = ?", seller.ID, domain.ProductPublished).Order("name asc").Find(&products).Error
	return seller, products, err
}

func (s *Catalog) Published(sellerSlug, productSlug string) (domain.Seller, domain.Product, error) {
	if domain.ReservedSlug(sellerSlug) || !domain.ValidSlug(sellerSlug) || !domain.ValidSlug(productSlug) {
		return domain.Seller{}, domain.Product{}, ErrNotFound
	}
	var seller domain.Seller
	if err := s.db.Where("slug = ?", sellerSlug).First(&seller).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Seller{}, domain.Product{}, ErrNotFound
		}
		return domain.Seller{}, domain.Product{}, err
	}
	var product domain.Product
	err := s.db.Where("seller_id = ? AND slug = ? AND status = ?", seller.ID, productSlug, domain.ProductPublished).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Seller{}, domain.Product{}, ErrNotFound
	}
	if err != nil {
		return domain.Seller{}, domain.Product{}, err
	}
	product.Seller = seller
	return seller, product, nil
}

func (s *Catalog) List(actor Actor) ([]domain.Product, error) {
	q := s.db.Preload("Seller").Preload("Category").Order("updated_at desc")
	if !actor.IsAdmin() {
		q = q.Where("seller_id = ?", actor.SellerID)
	}
	var products []domain.Product
	err := q.Find(&products).Error
	return products, err
}

func (s *Catalog) Get(actor Actor, id uint) (domain.Product, error) {
	var product domain.Product
	err := s.db.Preload("Seller").First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Product{}, ErrNotFound
	}
	if err != nil {
		return domain.Product{}, err
	}
	if !actor.IsAdmin() && product.SellerID != actor.SellerID {
		return domain.Product{}, ErrForbidden
	}
	return product, nil
}

func (s *Catalog) Save(actor Actor, product *domain.Product) error {
	if !actor.IsAdmin() {
		if actor.SellerID == 0 {
			return ValidationError{Message: "Profil toko belum ada"}
		}
		product.SellerID = actor.SellerID
	}
	cleaned, err := domain.ValidateProduct(*product)
	*product = cleaned
	if err != nil {
		return ValidationError{Message: err.Error()}
	}

	var seller domain.Seller
	if err := s.db.First(&seller, product.SellerID).Error; err != nil {
		return ValidationError{Message: "Toko tidak ditemukan"}
	}
	if domain.ReservedSlug(seller.Slug) {
		return ValidationError{Message: "Slug toko memakai jalur yang dicadangkan"}
	}
	var category domain.Category
	if err := s.db.First(&category, product.CategoryID).Error; err != nil {
		return ValidationError{Message: "Kategori tidak ditemukan"}
	}

	var count int64
	q := s.db.Model(&domain.Product{}).Where("seller_id = ? AND slug = ?", product.SellerID, product.Slug)
	if product.ID > 0 {
		q = q.Where("id <> ?", product.ID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrSlugTaken
	}

	if product.ID == 0 {
		return s.db.Omit("Seller", "Category").Create(product).Error
	}
	var existing domain.Product
	if err := s.db.First(&existing, product.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if !actor.IsAdmin() && existing.SellerID != actor.SellerID {
		return ErrForbidden
	}
	product.CreatedAt = existing.CreatedAt
	return s.db.Omit("Seller", "Category").Save(product).Error
}

func (s *Catalog) Delete(actor Actor, id uint) error {
	product, err := s.Get(actor, id)
	if err != nil {
		return err
	}
	return s.db.Delete(&domain.Product{}, product.ID).Error
}

func (s *Catalog) Categories() ([]domain.Category, error) {
	var categories []domain.Category
	err := s.db.Order("name asc").Find(&categories).Error
	return categories, err
}

func (s *Catalog) Sellers() ([]domain.Seller, error) {
	var sellers []domain.Seller
	err := s.db.Order("store_name asc").Find(&sellers).Error
	return sellers, err
}

func isDuplicate(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
