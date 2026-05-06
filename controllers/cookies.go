package controllers

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Metadata keys exchanged between gateway and gRPC handlers.
//
// Incoming (set by router.WithMetadata from HTTP cookies):
//
//	auth-access-token  — raw JWT
//	auth-refresh-token — opaque refresh token
//	user-agent         — client UA, captured for refresh-token records
//
// Outgoing (set by handlers via grpc.SetHeader, forwarded by router.WithOutgoingHeaderMatcher):
//
//	set-cookie — one or more Set-Cookie header values
const (
	MDAccessTokenKey  = "auth-access-token"
	MDRefreshTokenKey = "auth-refresh-token"
	MDUserAgentKey    = "user-agent"
)

const (
	CookieAccessToken  = "access_token"
	CookieRefreshToken = "refresh_token"
)

type CookieOptions struct {
	Domain string
	Secure bool
}

// SetAuthCookies sends Set-Cookie headers for both tokens to the gateway.
// Pass empty values + zero ttl to clear them (logout).
func SetAuthCookies(ctx context.Context, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, opts CookieOptions) error {
	md := metadata.MD{}

	if accessToken != "" {
		md.Append("set-cookie", buildCookie(CookieAccessToken, accessToken, accessTTL, "/v1", opts).String())
	} else {
		md.Append("set-cookie", buildCookie(CookieAccessToken, "", -time.Hour, "/v1", opts).String())
	}
	if refreshToken != "" {
		md.Append("set-cookie", buildCookie(CookieRefreshToken, refreshToken, refreshTTL, "/v1/auth", opts).String())
	} else {
		md.Append("set-cookie", buildCookie(CookieRefreshToken, "", -time.Hour, "/v1/auth", opts).String())
	}
	return grpc.SetHeader(ctx, md)
}

func buildCookie(name, value string, ttl time.Duration, path string, opts CookieOptions) *http.Cookie {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   opts.Secure,
	}
	if opts.Domain != "" {
		c.Domain = opts.Domain
	}
	if ttl > 0 {
		c.Expires = time.Now().Add(ttl)
		c.MaxAge = int(ttl.Seconds())
	} else {
		// Expire immediately.
		c.MaxAge = -1
		c.Expires = time.Unix(1, 0)
	}
	return c
}

// IncomingMetadataValue returns the first value for the given key from the
// incoming gRPC metadata, or empty string if not present.
func IncomingMetadataValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vs := md.Get(key)
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}
