package services

import (
	"errors"
	"strings"
	"time"

	db "AgnosAssignments/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	jwtSecret string
}

type CreateStaffInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Hospital string `json:"hospital" binding:"required"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

type StaffClaims struct {
	Username string `json:"username"`
	Hospital string `json:"hospital"`
	jwt.RegisteredClaims
}

func NewAuthService(database *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        database,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) CreateStaff(input CreateStaffInput) (*db.Staff, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Hospital = strings.TrimSpace(input.Hospital)

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	staff := db.Staff{
		Username:     input.Username,
		PasswordHash: string(hash),
		Hospital:     input.Hospital,
	}

	if err := s.db.Create(&staff).Error; err != nil {
		return nil, err
	}

	return &staff, nil
}

func (s *AuthService) Login(input LoginInput) (string, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Hospital = strings.TrimSpace(input.Hospital)

	var staff db.Staff
	if err := s.db.Where("username = ? AND hospital = ?", input.Username, input.Hospital).First(&staff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(input.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	claims := StaffClaims{
		Username: staff.Username,
		Hospital: staff.Hospital,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   staff.Username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(tokenString string) (*StaffClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &StaffClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*StaffClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) HasAnyStaff() (bool, error) {
	var count int64
	if err := s.db.Model(&db.Staff{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
