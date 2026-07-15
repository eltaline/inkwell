package user

import (
	"path/filepath"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("mypassword123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}

	u := &User{PasswordHash: hash}
	if !u.CheckPassword("mypassword123") {
		t.Error("CheckPassword should return true for correct password")
	}
	if u.CheckPassword("wrongpassword") {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestStoreCreateAndGet(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "users.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	u, err := s.Create("alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("Username = %q, want %q", u.Username, "alice")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "alice@example.com")
	}

	got, err := s.GetByUsername("alice")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("ID = %q, want %q", got.ID, u.ID)
	}

	got, err = s.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Username != "alice" {
		t.Errorf("Username = %q, want %q", got.Username, "alice")
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "users.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_, err = s.Create("alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = s.Create("alice", "alice2@example.com", "password456")
	if err != ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestStoreValidation(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "users.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  error
	}{
		{"empty username", "", "a@b.com", "password123", ErrEmptyUsername},
		{"invalid email", "bob", "notanemail", "password123", ErrInvalidEmail},
		{"weak password", "bob", "bob@example.com", "short", ErrWeakPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Create(tt.username, tt.email, tt.password)
			if err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "users.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_, err = s.GetByUsername("nobody")
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	_, err = s.GetByID("999")
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestStorePersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	s1, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_, err = s1.Create("alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Reopen store, user should still be there
	s2, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore reopen: %v", err)
	}
	u, err := s2.GetByUsername("alice")
	if err != nil {
		t.Fatalf("GetByUsername after reopen: %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("Username = %q, want %q", u.Username, "alice")
	}
}
