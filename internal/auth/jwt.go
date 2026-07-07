package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTypeAccess = "access"

type Claims struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id,omitempty"`
	Typ      string `json:"typ"`
	jwt.RegisteredClaims
}

type TokenIssuer struct {
	secret    []byte
	accessTTL time.Duration
}

func NewTokenIssuer(secret string, accessTTL time.Duration) (*TokenIssuer, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 bytes")
	}
	return &TokenIssuer{secret: []byte(secret), accessTTL: accessTTL}, nil
}

func (t *TokenIssuer) IssueAccess(userID, email string, role Role, tenantID string) (token string, jti string, expiresIn int, err error) {
	now := time.Now().UTC()
	exp := now.Add(t.accessTTL)
	jti = newJTI()
	claims := Claims{
		Email:    email,
		Role:     string(role),
		TenantID: tenantID,
		Typ:      tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString(t.secret)
	if err != nil {
		return "", "", 0, err
	}
	return signed, jti, int(t.accessTTL.Seconds()), nil
}

func (t *TokenIssuer) ParseAccess(tokenString string) (AuthContext, time.Time, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return t.secret, nil
	})
	if err != nil {
		return AuthContext{}, time.Time{}, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return AuthContext{}, time.Time{}, fmt.Errorf("invalid token")
	}
	if claims.Typ != tokenTypeAccess {
		return AuthContext{}, time.Time{}, fmt.Errorf("invalid token type")
	}
	role := Role(claims.Role)
	if !role.Valid() {
		return AuthContext{}, time.Time{}, fmt.Errorf("invalid role")
	}
	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	return AuthContext{
		UserID:   claims.Subject,
		Email:    claims.Email,
		Role:     role,
		TenantID: claims.TenantID,
		JTI:      claims.ID,
	}, exp, nil
}

func newJTI() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "jti_" + hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return "jti_" + hex.EncodeToString(b[:])
}