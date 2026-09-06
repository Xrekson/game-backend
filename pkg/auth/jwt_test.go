package auth

import (
	"testing"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-key-12345"
	mgr := NewJWTManager(secret, 1)

	token, err := mgr.GenerateToken("user-123", "testuser", "player")
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("expected token to be valid, got %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected UserID user-123, got %s", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("expected Username testuser, got %s", claims.Username)
	}
	if claims.Role != "player" {
		t.Errorf("expected Role player, got %s", claims.Role)
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	mgr := NewJWTManager("secret", 1)
	_, err := mgr.ValidateToken("invalid.token.string")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
