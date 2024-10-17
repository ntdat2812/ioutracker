package util

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JWTHelper struct {
	Secret               []byte
	RefreshSecret        []byte
	TokenExpireIn        time.Duration
	RefreshTokenExpireIn time.Duration
}

func NewJWTHelper() *JWTHelper {
	return &JWTHelper{
		Secret:               []byte(os.Getenv("JWT_SECRET")),
		RefreshSecret:        []byte(os.Getenv("JWT_REFRESH_SECRET")),
		TokenExpireIn:        time.Minute * 30,
		RefreshTokenExpireIn: time.Hour * 72,
	}
}

func (j *JWTHelper) GenerateToken(userId primitive.ObjectID, isRefreshToken bool) (string, error) {

	key := j.Secret
	expireIn := j.TokenExpireIn
	if isRefreshToken {
		key = j.RefreshSecret
		expireIn = j.RefreshTokenExpireIn
	}

	claims := jwt.MapClaims{
		"id":  userId.Hex(),
		"exp": time.Now().Add(expireIn).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func (j *JWTHelper) ParseWithClaims(jwtToken string, isRefreshToken bool) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	secret := j.Secret
	if isRefreshToken {
		secret = j.RefreshSecret
	}

	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
