package auth

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/eltaline/inkwell/internal/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

const (
	defaultTokenTTL = 24 * time.Hour
)

// Claims extends standard JWT claims with user-specific fields.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	Username string `json:"usr"`
}

// Service handles authentication: registration, login, token validation, and logout.
type Service struct {
	users    *user.Store
	secret   []byte
	tokenTTL time.Duration

	// revoked tracks logged-out token IDs until they expire naturally.
	revoked   map[string]time.Time
	revokedMu sync.RWMutex
}

// NewService creates an auth service with the given user store and JWT signing secret.
func NewService(users *user.Store, secret string) *Service {
	return &Service{
		users:    users,
		secret:   []byte(secret),
		tokenTTL: defaultTokenTTL,
		revoked:  make(map[string]time.Time),
	}
}

// Register creates a new user account and returns a JWT token.
func (s *Service) Register(username, email, password string) (string, error) {
	u, err := s.users.Create(username, email, password)
	if err != nil {
		return "", err
	}

	return s.generateToken(u)
}

// Login authenticates a user and returns a JWT token.
func (s *Service) Login(username, password string) (string, error) {
	u, err := s.users.GetByUsername(username)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !u.CheckPassword(password) {
		return "", ErrInvalidCredentials
	}

	return s.generateToken(u)
}

// ValidateToken parses and validates a JWT token string, returning the claims.
func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	s.revokedMu.RLock()
	_, revoked := s.revoked[claims.ID]
	s.revokedMu.RUnlock()

	if revoked {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

// Logout revokes the given token so it can no longer be used.
func (s *Service) Logout(tokenStr string) error {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return err
	}

	s.revokedMu.Lock()
	s.revoked[claims.ID] = claims.ExpiresAt.Time
	s.revokedMu.Unlock()

	s.cleanupRevoked()
	return nil
}

func (s *Service) generateToken(u *user.User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			Subject:   u.ID,
		},
		UserID:   u.ID,
		Username: u.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) cleanupRevoked() {
	now := time.Now()
	s.revokedMu.Lock()
	defer s.revokedMu.Unlock()

	for id, exp := range s.revoked {
		if now.After(exp) {
			delete(s.revoked, id)
		}
	}
}

func generateJTI() string {
	// Use timestamp + counter for uniqueness; no need for crypto/rand
	// since JTI only needs to be unique, not unpredictable.
	return time.Now().Format("20060102150405.000000000")
}
