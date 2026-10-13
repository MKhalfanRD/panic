package service

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"panic/internal/domain"
)

type Auth struct {
	db     *gorm.DB
	secret []byte
	ttl    time.Duration
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"uid"`
}

func NewAuth(db *gorm.DB, secret string, hours int) *Auth {
	if hours <= 0 {
		hours = 24
	}
	return &Auth{db: db, secret: []byte(secret), ttl: time.Duration(hours) * time.Hour}
}

func (s *Auth) Login(email, password string) (string, Actor, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var user domain.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", Actor{}, ErrInvalidLogin
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", Actor{}, ErrInvalidLogin
	}
	actor, err := s.actorFor(user)
	if err != nil {
		return "", Actor{}, err
	}
	token, err := s.sign(user.ID)
	if err != nil {
		return "", Actor{}, err
	}
	return token, actor, nil
}

func (s *Auth) ActorFromToken(token string) (Actor, error) {
	claims := &jwtClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid || claims.UserID == 0 {
		return Actor{}, ErrInvalidLogin
	}
	var user domain.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return Actor{}, ErrInvalidLogin
	}
	return s.actorFor(user)
}

func (s *Auth) actorFor(user domain.User) (Actor, error) {
	actor := Actor{User: user}
	var seller domain.Seller
	err := s.db.Where("user_id = ?", user.ID).First(&seller).Error
	if err == nil {
		actor.SellerID = seller.ID
	} else if err != gorm.ErrRecordNotFound {
		return Actor{}, err
	}
	return actor, nil
}

func (s *Auth) sign(userID uint) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
