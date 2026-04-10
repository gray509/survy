package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Hash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "survy",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
	})

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,             // the JWT string (e.g., "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
		&jwt.RegisteredClaims{}, // pointer to claims struct to decode into
		func(t *jwt.Token) (interface{}, error) {
			// return the key used to verify the token
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	userString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(userString)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil

}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("there is no authorization token")
	}
	token_string := strings.Split(auth, " ")
	if len(token_string) < 2 || token_string[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return token_string[1], nil

}

func MakeRefreshToken() (string, string, *time.Time, *time.Time, error) {
	token := make([]byte, 32)
	rand.Read(token)
	hexToken := hex.EncodeToString(token)
	hash, err := Hash(hexToken)
	if err != nil {
		return "", "", nil, nil, err
	}
	create_at := time.Now()
	expires_at := create_at.Add(time.Hour * 24 * 2)
	return hexToken, hash, &create_at, &expires_at, nil
}
