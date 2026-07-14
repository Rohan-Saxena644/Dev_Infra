package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

func HashPassword(password string) (string, error) {
	bytesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytesPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type Claims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(user database.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	secret := os.Getenv("SECRET")

	if secret == "" {
		return "", errors.New("SECRET environment variable is not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("SECRET")

	if secret == "" {
		return nil, errors.New("SECRET environment variable is not set")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *ProjectService) SignUp(email, password string) (string, database.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if !IsValidGmailID(email) {
		return "", database.User{}, errors.New("invalid gmail id")
	}
	if len(password) < 8 || len(password) > 72 {
		return "", database.User{}, errors.New("password must be between 8 and 72 characters")
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", database.User{}, err
	}

	user, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})

	if err != nil {
		return "", database.User{}, err
	}

	token, err := GenerateToken(user)

	if err != nil {
		return "", database.User{}, err
	}

	return token, user, nil
}

func (s *ProjectService) Login(email, password string) (string, database.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if !IsValidGmailID(email) {
		return "", database.User{}, errors.New("invalid gmail id")
	}

	user, err := s.DB.GetUserByEmail(context.Background(), email)
	if err != nil {
		return "", database.User{}, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return "", database.User{}, errors.New("invalid password")
	}

	token, err := GenerateToken(user)
	if err != nil {
		return "", database.User{}, err
	}

	return token, user, nil
}

func IsValidGmailID(id string) bool {
	// Parse the email address to check for valid RFC 5322 format
	parsedEmail, err := mail.ParseAddress(id)
	if err != nil {
		return false
	}
	if strings.ToLower(parsedEmail.Address) != strings.ToLower(id) {
		return false
	}

	// Extract the domain part
	parts := strings.Split(parsedEmail.Address, "@")
	if len(parts) != 2 {
		return false
	}

	// Verify the domain is specifically gmail.com
	return strings.ToLower(parts[1]) == "gmail.com"
}
