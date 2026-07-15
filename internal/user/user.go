package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")
	ErrInvalidEmail  = errors.New("invalid email")
	ErrWeakPassword  = errors.New("password must be at least 8 characters")
	ErrEmptyUsername = errors.New("username must not be empty")
)

// User represents a registered user.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CheckPassword compares a plaintext password against the stored hash.
func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// HashPassword returns a bcrypt hash of the given password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Store provides thread-safe user persistence backed by a JSON file.
type Store struct {
	path  string
	mu    sync.RWMutex
	users map[string]*User // keyed by username
	nextID int
}

// NewStore creates a Store that reads/writes users to the given JSON file.
// If the file exists, it loads existing users; otherwise it starts empty.
func NewStore(path string) (*Store, error) {
	s := &Store{
		path:  path,
		users: make(map[string]*User),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("user store: load: %w", err)
	}
	return s, nil
}

// Create adds a new user with the given credentials. It validates input,
// hashes the password, and persists to disk.
func (s *Store) Create(username, email, password string) (*User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))

	if username == "" {
		return nil, ErrEmptyUsername
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return nil, ErrUserExists
	}

	s.nextID++
	now := time.Now().UTC()
	u := &User{
		ID:           fmt.Sprintf("%d", s.nextID),
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.users[username] = u
	if err := s.save(); err != nil {
		delete(s.users, username)
		return nil, err
	}

	return u, nil
}

// GetByUsername returns the user with the given username or ErrUserNotFound.
func (s *Store) GetByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// GetByID returns the user with the given ID or ErrUserNotFound.
func (s *Store) GetByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var users []*User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("user store: unmarshal: %w", err)
	}

	maxID := 0
	for _, u := range users {
		s.users[u.Username] = u
		var id int
		if _, err := fmt.Sscanf(u.ID, "%d", &id); err == nil && id > maxID {
			maxID = id
		}
	}
	s.nextID = maxID
	return nil
}

func (s *Store) save() error {
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("user store: marshal: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("user store: write: %w", err)
	}
	return os.Rename(tmp, s.path)
}
