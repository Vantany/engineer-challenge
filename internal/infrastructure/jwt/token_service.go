package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/usecase"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrKeyParse     = errors.New("failed to parse rsa key")
)

type Config struct {
	PrivateKeyPath  string
	PublicKeyPath   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type TokenService struct {
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenService(cfg Config) (*TokenService, error) {
	if cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" {
		return nil, errors.New("jwt private and public key paths are required")
	}

	privKeyBytes, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	privKey, err := jwtlib.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return nil, ErrKeyParse
	}

	pubKeyBytes, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}
	pubKey, err := jwtlib.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, ErrKeyParse
	}

	accessTTL := cfg.AccessTokenTTL
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL := cfg.RefreshTokenTTL
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	return &TokenService{
		privateKey:      privKey,
		publicKey:       pubKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}, nil
}

type accessClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	jwtlib.RegisteredClaims
}

func (s *TokenService) GenerateAccessToken(user *domain.User, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(s.accessTokenTTL)

	claims := accessClaims{
		UserID: string(user.ID),
		Email:  string(user.Email),
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   string(user.ID),
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *TokenService) GenerateRefreshToken(now time.Time) (string, time.Time, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, err
	}

	token := hex.EncodeToString(tokenBytes)
	expiresAt := now.Add(s.refreshTokenTTL)
	return token, expiresAt, nil
}

func (s *TokenService) HashRefreshToken(token string) (string, error) {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func (s *TokenService) ValidateAccessToken(token string) (*usecase.TokenClaims, error) {
	parsed, err := jwtlib.ParseWithClaims(token, &accessClaims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return &usecase.TokenClaims{
		UserID:    domain.UserID(claims.UserID),
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
