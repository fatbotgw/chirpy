package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	//  Hash the password using the argon2id.CreateHash function
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	// Use the argon2id.ComparePasswordAndHash function to compare the 
	// password that the user entered in the HTTP request with the password 
	// that is stored in the database
	checkResult, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return checkResult, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy-access",
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))

	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
	    return []byte(tokenSecret), nil
	}

	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	// This function should look for the Authorization header in the headers 
	// parameter and return the TOKEN_STRING if it exists (stripping off the 
	// Bearer prefix and whitespace). 
	// 
	// Bearer TOKEN_STRING
	// 
	// If the header doesn't exist, return an error.
	tokenString := headers.Get("Authorization")
	if tokenString == "" {
		return "", errors.New("No authorization token")
	}
	token := strings.TrimPrefix(tokenString, "Bearer ")
	token = strings.TrimSpace(token)

	return token, nil
}
