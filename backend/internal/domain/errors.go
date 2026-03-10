package domain

import "errors"

var (
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrInvalidEmailFormat   = errors.New("invalid email format")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrResetTokenExpired    = errors.New("reset token expired")
	ErrResetTokenUsed       = errors.New("reset token already used")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrSessionNotFound      = errors.New("session not found")
	ErrPasswordTooWeak      = errors.New("password does not meet policy")
	ErrTooManyResetRequests = errors.New("too many reset requests")
)

