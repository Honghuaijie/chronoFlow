package service

import (
	"context"
	"strings"
	"time"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/auth"
	"chronoFlow-admin/internal/conf"
	httpErrors "chronoFlow-admin/internal/errors"
)

type AuthService struct {
	v1.UnimplementedAuthServer

	security *conf.Security
}

func NewAuthService(security *conf.Security) *AuthService {
	return &AuthService{security: security}
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	username := strings.TrimSpace(req.GetUsername())
	password := strings.TrimSpace(req.GetPassword())
	if username == "" || password == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "username 和 password 不能为空")
	}
	if s.security == nil || username != s.security.AdminUsername || password != s.security.AdminPassword {
		return nil, httpErrors.EWithMessage(httpErrors.ErrUnauthorized, "账号或密码错误")
	}
	token, err := auth.GenerateJWT(s.security.JwtSecret, auth.JWTClaims{UserID: 1}, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Code:    0,
		Message: successMessage("Login"),
		Data:    &v1.LoginReply_Data{Token: token, Username: username},
	}, nil
}

func (s *AuthService) Current(ctx context.Context, _ *v1.CurrentRequest) (*v1.CurrentReply, error) {
	username := ""
	if s.security != nil {
		username = s.security.AdminUsername
	}
	return &v1.CurrentReply{
		Code:    0,
		Message: successMessage("Current"),
		Data:    &v1.CurrentReply_Data{UserId: 1, Username: username, Role: "admin"},
	}, nil
}
