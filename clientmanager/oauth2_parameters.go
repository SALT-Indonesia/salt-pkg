package clientmanager

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type OAuth2Argument interface {
	string | *oauth2.Config
}

type OAuth2Parameters[T OAuth2Argument] struct {
	Auth             T      // required. Access token (string) or *oauth2.Config
	CodeFromCallback string // optional
}

func (p OAuth2Parameters[T]) Client() (*http.Client, error) {
	var token *oauth2.Token
	switch cred := any(p.Auth).(type) {
	case *oauth2.Config:
		var err error
		token, err = cred.Exchange(context.Background(), p.CodeFromCallback)
		if err != nil {
			return nil, err
		}
	case string:
		token = &oauth2.Token{AccessToken: cred}
	}
	return oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token)), nil
}
