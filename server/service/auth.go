package service

import (
	"fmt"
	"net/http"
	"time"
	"errors"

	"github.com/dgrijalva/jwt-go"

	"github.com/sergivb01/forums/user"
)

var errNotLoggedIn = errors.New("you are not logged in")

func (s *Server) getClaims(r *http.Request) (*claims, error) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, errNotLoggedIn
		}
		return nil, err
	}

	rawToken := c.Value

	claim := &claims{}

	tkn, err := jwt.ParseWithClaims(rawToken, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("coulnd't parse your token: %v", err)
	}

	if !tkn.Valid {
		return nil, fmt.Errorf("invalid %s JWT token", tkn.Raw)
	}

	expires := time.Unix(claim.ExpiresAt, 0)

	// check if the code has expired
	if time.Now().After(expires) {
		return nil, fmt.Errorf("your token expired at %s, %s ago", expires, time.Since(expires))
	}

	return claim, nil
}

type claims struct {
	User user.User `json:"user"`
	jwt.StandardClaims
}

type WebToken struct {
	TokenString string    `json:"token"`
	ExpiresAt   time.Time `json:"expirationTime"`
}

func (s *Server) generateToken(user *user.User) (*WebToken, error) {
	expirationTime := time.Now().Add(s.cfg.JWT.Duration)

	claim := &claims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)

	// Create the JWT string
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &WebToken{
		TokenString: tokenString,
		ExpiresAt:   expirationTime,
	}, nil
}
