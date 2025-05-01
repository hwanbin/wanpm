package main

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	Issuer        string
	Audience      string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

type JwtUser struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func readPrivateKey() (*rsa.PrivateKey, error) {
	privateKeyFile, err := os.Open("private_key.pem")
	if err != nil {
		return nil, err
	}
	defer privateKeyFile.Close()

	keyBytes, err := io.ReadAll(privateKeyFile)
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func readPublicKey() (*rsa.PublicKey, error) {
	publicKeyFile, err := os.Open("public_key.pem")
	if err != nil {
		return nil, err
	}
	defer publicKeyFile.Close()

	keyBytes, err := io.ReadAll(publicKeyFile)
	if err != nil {
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func (j *Auth) GenerateRSAedTokenPair(user *JwtUser) (TokenPairs, error) {
	// func (j *Auth) GenerateRSAedTokenPair(user *data.User) (TokenPairs, error) {
	privateKey, err := readPrivateKey()
	if err != nil {
		return TokenPairs{}, nil
	}

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = user.Email
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	signedAccessToken, err := token.SignedString(privateKey)
	if err != nil {
		return TokenPairs{}, err
	}

	refreshToken := jwt.New(jwt.SigningMethodRS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = user.Email
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	refreshTokenClaims["exp"] = time.Now().Add(j.RefreshExpiry).Unix()

	signedRefreshToken, err := refreshToken.SignedString(privateKey)
	if err != nil {
		return TokenPairs{}, err
	}

	var TokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	return TokenPairs, nil
}

func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:        j.CookieName,
		Path:        j.CookiePath,
		Value:       refreshToken,
		Expires:     time.Now().Add(j.RefreshExpiry),
		MaxAge:      int(j.RefreshExpiry.Seconds()),
		SameSite:    http.SameSiteNoneMode,
		Domain:      j.CookieDomain,
		Partitioned: true,
		HttpOnly:    true,
		Secure:      true,
	}
}

func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:        j.CookieName,
		Path:        j.CookiePath,
		Value:       "",
		Expires:     time.Unix(0, 0),
		MaxAge:      -1,
		SameSite:    http.SameSiteNoneMode,
		Domain:      j.CookieDomain,
		Partitioned: true,
		HttpOnly:    true,
		Secure:      true,
	}
}

func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	w.Header().Add("Vary", "Authorization")

	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("invalid auth header")
	}

	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header")
	}

	token := headerParts[1]

	claims := &Claims{}

	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			publicKey, err := readPublicKey()
			if err != nil {
				return nil, err
			}

			return publicKey, nil
		},
	)

	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}

	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}

	return token, claims, nil
}
