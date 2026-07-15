package auth

import (
	"path/filepath"
	"testing"

	"github.com/eltaline/inkwell/internal/user"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	dir := t.TempDir()
	users, err := user.NewStore(filepath.Join(dir, "users.json"))
	if err != nil {
		t.Fatalf("user store: %v", err)
	}
	return NewService(users, "test-secret-key-for-testing")
}

func TestRegisterAndLogin(t *testing.T) {
	svc := newTestService(t)

	token, err := svc.Register("alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if token == "" {
		t.Fatal("Register returned empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Username != "alice" {
		t.Errorf("Username = %q, want %q", claims.Username, "alice")
	}

	loginToken, err := svc.Login("alice", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loginToken == "" {
		t.Fatal("Login returned empty token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Register("bob", "bob@example.com", "correctpass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, err = svc.Login("bob", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginNonexistent(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Login("nobody", "password123")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	svc := newTestService(t)

	token, err := svc.Register("alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.Logout(token); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err != ErrTokenRevoked {
		t.Fatalf("expected ErrTokenRevoked, got %v", err)
	}
}

func TestInvalidToken(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.ValidateToken("not.a.valid.token")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
