package domain

import "time"

const (
	RoleAdmin  = "admin"
	RoleSeller = "seller"

	ProductDraft     = "draft"
	ProductPublished = "published"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:20;not null;index" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Seller struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	StoreName   string    `gorm:"size:120;not null" json:"store_name"`
	Slug        string    `gorm:"size:80;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"size:500" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	Slug      string    `gorm:"size:80;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SellerID      uint      `gorm:"not null;uniqueIndex:idx_seller_product_slug" json:"seller_id"`
	Seller        Seller    `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	CategoryID    uint      `gorm:"not null;index" json:"category_id"`
	Category      Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Name          string    `gorm:"size:160;not null" json:"name"`
	Slug          string    `gorm:"size:80;not null;uniqueIndex:idx_seller_product_slug" json:"slug"`
	Price         int64     `gorm:"not null" json:"price"`
	OriginalPrice int64     `gorm:"not null;default:0" json:"original_price"`
	Status        string    `gorm:"size:20;not null;default:draft;index" json:"status"`
	ImageURL      string    `gorm:"type:text" json:"image_url,omitempty"`
	Landing       Landing   `gorm:"serializer:json;type:longtext" json:"landing"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Order struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Invoice         string     `gorm:"size:40;uniqueIndex;not null" json:"invoice"`
	SellerID        uint       `gorm:"not null;index" json:"seller_id"`
	Seller          Seller     `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	ProductID       uint       `gorm:"not null;index" json:"product_id"`
	ProductName     string     `gorm:"size:160;not null" json:"product_name"`
	CustomerName    string     `gorm:"size:100;not null" json:"customer_name"`
	CustomerEmail   string     `gorm:"size:255;not null" json:"customer_email"`
	CustomerPhone   string     `gorm:"size:30;not null" json:"customer_phone"`
	GrossAmount     int64      `gorm:"not null" json:"gross_amount"`
	Status          string     `gorm:"size:20;not null;index" json:"status"`
	MidtransOrderID string     `gorm:"size:64;index" json:"midtrans_order_id,omitempty"`
	SnapToken       string     `gorm:"type:text" json:"-"`
	SnapRedirectURL string     `gorm:"type:text" json:"-"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
