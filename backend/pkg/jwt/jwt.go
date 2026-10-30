package jwt

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrExpired = errors.New("token đã hết hạn")
	ErrInvalid = errors.New("token không hợp lệ")
)

type Claims struct {
	UserID      uuid.UUID `json:"uid"`
	EmployeeID  uuid.UUID `json:"eid"`
	SessionID   uuid.UUID `json:"sid"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"perms"`
	Scope       string    `json:"scope"`
	Org         string    `json:"org,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret    []byte
	accessTTL time.Duration
	issuer    string
}

func NewManager(secret string, accessTTL time.Duration, issuer string) (*Manager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET phải dài ít nhất 32 ký tự, hiện có %d", len(secret))
	}
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, issuer: issuer}, nil
}

func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

func (m *Manager) Issue(c Claims) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTTL)

	c.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   c.UserID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        uuid.New().String(),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("ký token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *Manager) Verify(tokenString string) (*Claims, error) {
	var claims Claims

	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("thuật toán ký không hợp lệ: %v", t.Header["alg"])
		}
		return m.secret, nil
	},
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	return &claims, nil
}
