package domain

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordPolicy struct {
	MinLength int
	BcryptCost int
}

func NewDefaultPasswordPolicy() *BcryptPasswordPolicy {
	return &BcryptPasswordPolicy{
		MinLength:  8,
		BcryptCost: 12,
	}
}

func (p *BcryptPasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return ErrPasswordTooWeak
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrPasswordTooWeak
	}

	return nil
}

func (p *BcryptPasswordPolicy) Hash(password string) (string, error) {
	if err := p.Validate(password); err != nil {
		return "", err
	}

	cost := p.BcryptCost
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *BcryptPasswordPolicy) Verify(password, hash string) bool {
	// bcrypt.CompareHashAndPassword уже реализует защиту от timing attacks
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

