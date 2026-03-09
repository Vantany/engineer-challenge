package grpc

import (
	"context"

	"strings"

	"auth-service/internal/domain"
	authv1 "auth-service/internal/transport/grpc/gen/api/auth/v1"
	cmd "auth-service/internal/usecase/commands"
	qry "auth-service/internal/usecase/queries"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer

	RegisterHandler      *cmd.RegisterHandler
	LoginHandler         *cmd.LoginHandler
	LogoutHandler        *cmd.LogoutHandler
	RefreshHandler       *cmd.RefreshTokenHandler
	RequestResetHandler  *cmd.RequestPasswordResetHandler
	ResetPasswordHandler *cmd.ResetPasswordHandler
	ValidateTokenHandler *qry.ValidateTokenHandler
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	res, err := h.RegisterHandler.Handle(ctx, cmd.RegisterCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RegisterResponse{
		UserId:  string(res.User.ID),
		Session: mapSession(res.AccessToken, res.RefreshToken, res.AccessTokenExpiry.Unix(), res.User),
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	res, err := h.LoginHandler.Handle(ctx, cmd.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.LoginResponse{
		Session: mapSession(res.AccessToken, res.RefreshToken, res.AccessTokenExpiry.Unix(), res.User),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	err := h.LogoutHandler.Handle(ctx, cmd.LogoutCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.LogoutResponse{}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	res, err := h.RefreshHandler.Handle(ctx, cmd.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RefreshTokenResponse{
		Session: mapSession(res.AccessToken, res.RefreshToken, res.AccessTokenExpiry.Unix(), res.User),
	}, nil
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*authv1.RequestPasswordResetResponse, error) {
	err := h.RequestResetHandler.Handle(ctx, cmd.RequestPasswordResetCommand{
		Email: req.Email,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RequestPasswordResetResponse{}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	res, err := h.ResetPasswordHandler.Handle(ctx, cmd.ResetPasswordCommand{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.ResetPasswordResponse{
		Session: mapSession(res.AccessToken, res.RefreshToken, res.AccessTokenExpiry.Unix(), res.User),
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	accessToken := req.AccessToken

	if accessToken == "" {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
				accessToken = extractBearer(authHeaders[0])
			} else if authHeaders := md.Get("grpcgateway-authorization"); len(authHeaders) > 0 {
				accessToken = extractBearer(authHeaders[0])
			}
		}
	}

	res, err := h.ValidateTokenHandler.Handle(ctx, qry.ValidateTokenQuery{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, mapError(err)
	}

	if !res.Valid {
		return &authv1.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid:     true,
		User:      mapUser(res.User),
		ExpiresAt: res.ExpiresAt.Unix(),
	}, nil
}

func (h *AuthHandler) GetMe(ctx context.Context, req *authv1.GetMeRequest) (*authv1.GetMeResponse, error) {
	var accessToken string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			accessToken = extractBearer(authHeaders[0])
		} else if authHeaders := md.Get("grpcgateway-authorization"); len(authHeaders) > 0 {
			accessToken = extractBearer(authHeaders[0])
		}
	}

	if accessToken == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	res, err := h.ValidateTokenHandler.Handle(ctx, qry.ValidateTokenQuery{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, mapError(err)
	}

	if !res.Valid || res.User == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &authv1.GetMeResponse{
		User: mapUser(res.User),
	}, nil
}

func extractBearer(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func mapUser(u *domain.User) *authv1.User {
	if u == nil {
		return nil
	}
	return &authv1.User{
		Id:        string(u.ID),
		Email:     string(u.Email),
		CreatedAt: u.CreatedAt.Unix(),
	}
}

func mapSession(accessToken, refreshToken string, expiresAt int64, u *domain.User) *authv1.Session {
	return &authv1.Session{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         mapUser(u),
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch err {
	case domain.ErrInvalidEmailFormat, domain.ErrPasswordTooWeak:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrEmailAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrInvalidCredentials, domain.ErrResetTokenExpired, domain.ErrResetTokenUsed, domain.ErrSessionNotFound, domain.ErrSessionRevoked:
		return status.Error(codes.Unauthenticated, err.Error())
	case domain.ErrTooManyResetRequests:
		return status.Error(codes.ResourceExhausted, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
