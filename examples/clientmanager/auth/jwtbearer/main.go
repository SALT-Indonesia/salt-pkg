package main

import (
	"context"
	"log"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	auth := clientmanager.AuthJWT(
		"mysecretkey",
		jwt.SigningMethodHS256,
		clientmanager.AuthJWTClaims{
			Sub: "myusername",
			Iss: "myissuer",
			Aud: "myaudience",
			Nbf: time.Now(),
			Exp: time.Now().Add(time.Hour),
			Jti: clientmanager.AuthJWTClaimsJWTID{
				Generate: true,
			},
			Extra: map[string]any{
				"name": "John Doe",
			},
		},
	)

	res, err := clientmanager.Call[any](
		context.Background(),
		"https://httpbin.org/bearer",
		clientmanager.WithAuth(auth),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.IsSuccess())
}
