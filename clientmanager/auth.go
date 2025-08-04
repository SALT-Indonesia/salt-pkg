package clientmanager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hiyosi/hawk"
)

type Auth func(*http.Request) error

func AuthBasic(username, password string) Auth {
	return func(r *http.Request) error {
		r.SetBasicAuth(username, password)

		return nil
	}
}

func AuthBearer(token string) Auth {
	return func(r *http.Request) error {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		return nil
	}
}

func AuthAPIKey(key, value string, addToQueryParams bool) Auth {
	return func(r *http.Request) error {
		if addToQueryParams {
			urlValues := r.URL.Query()
			urlValues.Add(key, value)
			r.URL.RawQuery = urlValues.Encode()
		} else {
			r.Header.Add(key, value)
		}

		return nil
	}
}

// AuthJWT generates a JWT token to authorize an HTTP request.
//
// Parameters:
//   - secret: a plain secret. You need to decode it yourself if the secret is encoded.
//   - signingMethod: a signing method that is available on https://pkg.go.dev/github.com/golang-jwt/jwt/v5#SigningMethod
//   - claims: your claims here. It includes common claims, and you can provide custom claims if available.
//
// Returns:
//   - Auth: the authorization method
func AuthJWT(secret string, signingMethod jwt.SigningMethod, claims AuthJWTClaims) Auth {
	token := jwt.NewWithClaims(signingMethod, claims.mapClaims())
	signedToken, _ := token.SignedString([]byte(secret))

	return func(r *http.Request) error {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signedToken))

		return nil
	}
}

func AuthHawk(id, key string, option *hawk.Option) Auth {
	cred := &hawk.Credential{
		ID:  id,
		Key: key,
		Alg: hawk.SHA256,
	}
	if option == nil {
		option = &hawk.Option{}
	}
	client := hawk.NewClient(cred, option)

	return func(r *http.Request) error {
		header, _ := client.Header(r.Method, r.URL.String())

		r.Header.Add("Authorization", header)

		return nil
	}
}

func AuthAWS(params AWSParameters) Auth {
	return func(r *http.Request) error {
		return params.Signer(r)
	}
}

func AuthESB(apiKey, secret string) Auth {
	return func(r *http.Request) error {
		hasher := md5.New()
		hasher.Write(fmt.Appendf(
			[]byte{},
			"%s%s%d",
			apiKey,
			secret,
			time.Now().Unix(),
		))
		signature := hex.EncodeToString(hasher.Sum(nil))

		r.Header.Add("api_key", apiKey)
		r.Header.Add("x-signature", signature)

		return nil
	}
}
