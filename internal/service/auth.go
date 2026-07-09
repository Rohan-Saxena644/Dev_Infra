package service

import (
	"context"
	"github.com/Rohan-Saxena644/devinfra/internal/database"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"net/mail"
	"errors"
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


func (s *ProjectService) SignUp(email,password string) (database.User, error) {

	passwordHash, err := HashPassword(password)
	if err != nil {
		return database.User{}, err
	}

	if !IsValidGmailID(email) {
		return database.User{}, errors.New("invalid gmail id")
	}
	
	return s.DB.CreateUser(context.Background(), database.CreateUserParams{
		Email: email,
		PasswordHash: passwordHash,
	})
}



func IsValidGmailID(id string) bool {
	// Parse the email address to check for valid RFC 5322 format
	parsedEmail, err := mail.ParseAddress(id)
	if err != nil {
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