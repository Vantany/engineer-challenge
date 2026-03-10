package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserID string

type Email string

type AuthMethod struct {
	ID         string
	UserID     UserID
	Provider   string
	Credential string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type User struct {
	ID          UserID       `db:"id"`
	Email       Email        `db:"email"`
	AuthMethods []AuthMethod `db:"-"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

type PasswordPolicy interface {
	Validate(password string) error
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

func NewUser(id UserID, email Email, password string, policy PasswordPolicy, now time.Time) (*User, error) {
	if policy == nil {
		return nil, ErrPasswordTooWeak
	}

	if err := policy.Validate(password); err != nil {
		return nil, err
	}

	hash, err := policy.Hash(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Email:    email,
		AuthMethods: []AuthMethod{
			{
				ID:         uuid.NewString(),
				UserID:     id,
				Provider:   "password",
				Credential: hash,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *User) VerifyPassword(password string, policy PasswordPolicy) bool {
	if u == nil || policy == nil {
		return false
	}
	for _, am := range u.AuthMethods {
		if am.Provider == "password" {
			return policy.Verify(password, am.Credential)
		}
	}
	return false
}

func (u *User) UpdatePassword(newPassword string, policy PasswordPolicy, now time.Time) error {
	if policy == nil {
		return ErrPasswordTooWeak
	}
	if err := policy.Validate(newPassword); err != nil {
		return err
	}
	hash, err := policy.Hash(newPassword)
	if err != nil {
		return err
	}

	updated := false
	for i, am := range u.AuthMethods {
		if am.Provider == "password" {
			u.AuthMethods[i].Credential = hash
			u.AuthMethods[i].UpdatedAt = now
			updated = true
			break
		}
	}

	if !updated {
		u.AuthMethods = append(u.AuthMethods, AuthMethod{
			ID:         uuid.NewString(),
			UserID:     u.ID,
			Provider:   "password",
			Credential: hash,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	u.UpdatedAt = now
	return nil
}

func (u *User) CreateResetToken(rawToken string, now time.Time) *ResetToken {
	return &ResetToken{
		Token:     rawToken,
		UserID:    u.ID,
		ExpiresAt: now.Add(DefaultResetTokenTTL),
		Used:      false,
	}
}
