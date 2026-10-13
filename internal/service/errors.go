package service

import "errors"

var (
	ErrNotFound       = errors.New("tidak ditemukan")
	ErrForbidden      = errors.New("tidak diizinkan")
	ErrSlugTaken      = errors.New("slug sudah dipakai di toko ini")
	ErrNotConfigured  = errors.New("midtrans belum diatur")
	ErrBadSignature   = errors.New("signature midtrans tidak valid")
	ErrAmountMismatch = errors.New("nominal tidak sesuai")
	ErrInvalidLogin   = errors.New("email atau kata sandi salah")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
