package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidation(t *testing.T) {
    // arrange, act, assert
	// Arrange — set up inputs (a user ID, a secret, an expiration)
	// Act — call MakeJWT and/or ValidateJWT
	// Assert — check the result matches your expectation

	// Valid token created and validated with the same secret

	// Returns the original user ID, no error

	user := uuid.New()
	secret := "abcd1234"
	expiration := time.Hour

	newJWT, err := MakeJWT(user, secret, expiration)
	if err != nil {
		t.Fatalf("Error creating JWT: %s", err)
	}
	valUUID, err := ValidateJWT(newJWT, secret)
	if err != nil {
		t.Fatalf("Error validating JWT: %s", err)
	}
	if valUUID != user {
		t.Errorf("Input %v, Validation returned: %v", user, valUUID)
	}
}

func TestWrongSecret(t *testing.T) {
	user := uuid.New()
	secretGood := "abcd1234"
	secretBad  := "xyz789"
	expiration := time.Hour

	newJWT, err := MakeJWT(user, secretGood, expiration)
	if err != nil {
		t.Fatalf("Error creating JWT: %s", err)
	}
	_, err = ValidateJWT(newJWT, secretBad)
	if err == nil {
	    t.Errorf("Expected error but got none")
	}
}

func TestExpiredToken(t *testing.T) {
	user := uuid.New()
	secret := "abcd1234"
	// expiration := time.Hour
	expired := -time.Hour

	newJWT, err := MakeJWT(user, secret, expired)
	if err != nil {
		t.Fatalf("Error creating JWT: %s", err)
	}
	_, err = ValidateJWT(newJWT, secret)
	if err == nil {
	    t.Errorf("Expected error but got none")
	}
}