package jwt

import (
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/imagems/pkg/model"
	"github.com/dgrijalva/jwt-go"
)

type JWTer interface {
	Validate(token string, cs jwt.Claims) (*jwt.Token, error)
}

type Validator struct {
	errors.AuthErrCheck
	jwter JWTer
}

func NewValidator(jwter JWTer) (*Validator, error) {
	if jwter == nil {
		return nil, errors.Newf("JWTer was nil")
	}
	return &Validator{jwter: jwter}, nil
}

func (v Validator) Validate(token string) (*model.JWTClaim, error) {
	claim := &model.JWTClaim{}
	if _, err := v.jwter.Validate(token, claim); err != nil {
		return nil, errors.NewAuth(err)
	}
	return claim, nil
}
