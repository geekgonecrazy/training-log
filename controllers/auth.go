package controllers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/geekgonecrazy/training-log/config"
	"github.com/geekgonecrazy/training-log/core/auth"
	habitv1 "github.com/geekgonecrazy/training-log/models/habit/v1"
	"github.com/geekgonecrazy/training-log/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthController implements habitv1.AuthServiceServer.
type AuthController struct {
	habitv1.UnimplementedAuthServiceServer
	Store  store.Store
	Cfg    config.AuthConfig
	Cookie CookieOptions
	Now    func() time.Time
}

func NewAuthController(s store.Store, cfg config.AuthConfig, cookie CookieOptions) *AuthController {
	return &AuthController{Store: s, Cfg: cfg, Cookie: cookie, Now: time.Now}
}

func (c *AuthController) Register(ctx context.Context, req *habitv1.RegisterRequest) (*habitv1.RegisterResponse, error) {
	if !c.Cfg.RegistrationOpen {
		return nil, status.Error(codes.PermissionDenied, "registration is closed")
	}
	email := strings.TrimSpace(strings.ToLower(req.GetEmail()))
	if email == "" || !strings.Contains(email, "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}
	if len(req.GetPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	hash, err := auth.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "hash password: %v", err)
	}
	u, err := c.Store.Users().Create(ctx, email, hash, strings.TrimSpace(req.GetName()))
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
		return nil, status.Errorf(codes.Internal, "create user: %v", err)
	}
	return &habitv1.RegisterResponse{User: userToProto(u)}, nil
}

func (c *AuthController) Login(ctx context.Context, req *habitv1.LoginRequest) (*habitv1.LoginResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.GetEmail()))
	u, err := c.Store.Users().GetByEmail(ctx, email)
	if err != nil {
		// Don't leak whether the email exists.
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err := auth.VerifyPassword(u.PasswordHash, req.GetPassword()); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if err := c.issueTokens(ctx, u.ID, "", req.GetRememberMe()); err != nil {
		return nil, status.Errorf(codes.Internal, "issue tokens: %v", err)
	}
	return &habitv1.LoginResponse{User: userToProto(u)}, nil
}

func (c *AuthController) Refresh(ctx context.Context, _ *habitv1.RefreshRequest) (*habitv1.RefreshResponse, error) {
	tok := IncomingMetadataValue(ctx, MDRefreshTokenKey)
	if tok == "" {
		return nil, status.Error(codes.Unauthenticated, "no refresh token")
	}
	hash := auth.HashRefreshToken(tok)

	rec, err := c.Store.RefreshTokens().GetByHash(ctx, hash)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	now := c.Now()

	// Reuse of an already-revoked refresh token = potential theft. Revoke the entire family.
	if rec.RevokedAt != nil {
		_ = c.Store.RefreshTokens().RevokeFamily(ctx, rec.FamilyID, now)
		return nil, status.Error(codes.Unauthenticated, "refresh token reused; session terminated")
	}
	if !now.Before(rec.ExpiresAt) {
		return nil, status.Error(codes.Unauthenticated, "refresh token expired")
	}

	// Revoke the old, mint a new one in the same family.
	if err := c.Store.RefreshTokens().Revoke(ctx, rec.ID, now); err != nil {
		return nil, status.Errorf(codes.Internal, "revoke: %v", err)
	}
	if err := c.issueTokens(ctx, rec.UserID, rec.FamilyID, true); err != nil {
		return nil, status.Errorf(codes.Internal, "issue tokens: %v", err)
	}
	return &habitv1.RefreshResponse{}, nil
}

func (c *AuthController) Logout(ctx context.Context, _ *habitv1.LogoutRequest) (*habitv1.LogoutResponse, error) {
	tok := IncomingMetadataValue(ctx, MDRefreshTokenKey)
	if tok != "" {
		if rec, err := c.Store.RefreshTokens().GetByHash(ctx, auth.HashRefreshToken(tok)); err == nil {
			_ = c.Store.RefreshTokens().Revoke(ctx, rec.ID, c.Now())
		}
	}
	if err := SetAuthCookies(ctx, "", "", 0, 0, c.Cookie); err != nil {
		return nil, status.Errorf(codes.Internal, "clear cookies: %v", err)
	}
	return &habitv1.LogoutResponse{}, nil
}

func (c *AuthController) Me(ctx context.Context, _ *habitv1.MeRequest) (*habitv1.MeResponse, error) {
	uid, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "not signed in")
	}
	u, err := c.Store.Users().GetByID(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load user: %v", err)
	}
	return &habitv1.MeResponse{User: userToProto(u)}, nil
}

// issueTokens mints a new access JWT + refresh token, persists the refresh
// record (continuing the given familyID, or starting a new family if empty),
// and sets the cookies on the outgoing response.
func (c *AuthController) issueTokens(ctx context.Context, userID int64, familyID string, longSession bool) error {
	now := c.Now()
	access, err := auth.IssueAccessToken([]byte(c.Cfg.JWTSecret), userID, c.Cfg.AccessTokenTTL, now)
	if err != nil {
		return err
	}
	refresh, refreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return err
	}
	if familyID == "" {
		familyID, err = auth.NewFamilyID()
		if err != nil {
			return err
		}
	}

	refreshTTL := c.Cfg.RefreshTokenTTL
	if !longSession {
		// "Don't remember me" → cap to one day.
		if refreshTTL > 24*time.Hour {
			refreshTTL = 24 * time.Hour
		}
	}

	if _, err := c.Store.RefreshTokens().Create(ctx, &store.RefreshToken{
		UserID:    userID,
		TokenHash: refreshHash,
		FamilyID:  familyID,
		ExpiresAt: now.Add(refreshTTL),
		CreatedAt: now,
		UserAgent: IncomingMetadataValue(ctx, MDUserAgentKey),
	}); err != nil {
		return err
	}

	return SetAuthCookies(ctx, access, refresh, c.Cfg.AccessTokenTTL, refreshTTL, c.Cookie)
}
