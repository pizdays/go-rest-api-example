package util

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

// ParseToken parse JWT token string.
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, fmt.Errorf("util.ParseToken: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("util.ParseToken: token is invalid")
}

// CreateJwtToken generates JWT token with user ID, expiry date, and type in
// payload.
func CreateJwtToken(usrID uint, exp int64, signingMethod jwt.SigningMethod, key, tokenType string) (string, error) {

	claims := jwt.MapClaims{
		"id":   usrID,
		"type": tokenType,
	}

	if exp != 0 { //support log live token
		claims["exp"] = exp
	}

	tok := jwt.NewWithClaims(signingMethod, claims)

	tokStr, err := tok.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("util.CreateJwtToken: %w", err)
	}

	return tokStr, nil
}

// CreateAccessToken generates access jwt token.
func CreateAccessToken(userID uint, expiredAt int64) (string, error) {
	return CreateJwtToken(userID,
		expiredAt,
		jwt.SigningMethodHS256,
		os.Getenv("JWT_SECRET_KEY"),
		"access",
	)
}

// CreateRefreshToken generates refresh jwt token.
func CreateRefreshToken(userID uint, expiredAt int64) (string, error) {
	return CreateJwtToken(userID, expiredAt,
		jwt.SigningMethodHS256,
		os.Getenv("JWT_SECRET_KEY"),
		"refresh",
	)
}
