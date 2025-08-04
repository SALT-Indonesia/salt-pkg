package clientmanager

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthJWTClaimsJWTID struct {
	Generate bool   // generate the UUID for the JWT ID
	Custom   string // your custom JWT ID if available
}

type AuthJWTClaims struct {
	Sub   string             // subject of the JWT (the user)
	Iss   string             // issuer of the JWT
	Aud   string             // recipient for which the JWT is intended
	Nbf   time.Time          // time before which the JWT must not be accepted for processing
	Exp   time.Time          // time after which the JWT expires
	Jti   AuthJWTClaimsJWTID // unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	Extra map[string]any     // your custom claims here
}

func (a AuthJWTClaims) mapClaims() jwt.MapClaims {
	mapClaims := jwt.MapClaims{
		"iat": time.Now().Unix(),
	}
	if a.Sub != "" {
		mapClaims["sub"] = a.Sub
	}
	if a.Iss != "" {
		mapClaims["iss"] = a.Iss
	}
	if a.Aud != "" {
		mapClaims["aud"] = a.Aud
	}
	if !a.Nbf.IsZero() {
		mapClaims["nbf"] = a.Nbf.Unix()
	}
	if !a.Exp.IsZero() {
		mapClaims["exp"] = a.Exp.Unix()
	}
	if a.Jti.Custom != "" {
		mapClaims["jti"] = a.Jti.Custom
	} else if a.Jti.Generate {
		mapClaims["jti"] = uuid.New()
	}
	for k, v := range a.Extra {
		mapClaims[k] = v
	}
	return mapClaims
}
