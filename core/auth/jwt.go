package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Minimal JWT (HS256) issue/parse. We do this by hand to avoid pulling another dep.

var (
	ErrJWTMalformed = errors.New("jwt malformed")
	ErrJWTBadSig    = errors.New("jwt bad signature")
	ErrJWTExpired   = errors.New("jwt expired")
)

type Claims struct {
	Subject  int64 // user_id
	IssuedAt time.Time
	Expires  time.Time
}

func IssueAccessToken(secret []byte, userID int64, ttl time.Duration, now time.Time) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	payload := map[string]any{
		"sub": strconv.FormatInt(userID, 10),
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}
	payloadJSON, _ := json.Marshal(payload)

	hb := base64.RawURLEncoding.EncodeToString(headerJSON)
	pb := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signing := hb + "." + pb

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signing))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signing + "." + sig, nil
}

func ParseAccessToken(secret []byte, token string, now time.Time) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrJWTMalformed
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	want := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(parts[2])) {
		return nil, ErrJWTBadSig
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrJWTMalformed
	}
	var raw struct {
		Sub string `json:"sub"`
		Iat int64  `json:"iat"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &raw); err != nil {
		return nil, ErrJWTMalformed
	}
	uid, err := strconv.ParseInt(raw.Sub, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse sub: %w", err)
	}
	exp := time.Unix(raw.Exp, 0)
	if !now.Before(exp) {
		return nil, ErrJWTExpired
	}
	return &Claims{
		Subject:  uid,
		IssuedAt: time.Unix(raw.Iat, 0),
		Expires:  exp,
	}, nil
}
