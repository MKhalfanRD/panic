package service

import "panic/internal/domain"

type Actor struct {
	User     domain.User
	SellerID uint
}

func (a Actor) IsAdmin() bool {
	return a.User.Role == domain.RoleAdmin
}
