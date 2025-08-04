package clientmanager

import (
	"net/http"

	"github.com/dghubble/oauth1"
)

type OAuth1Parameters struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	TokenSecret    string
}

func (p OAuth1Parameters) Client() *http.Client {
	config := oauth1.NewConfig(p.ConsumerKey, p.ConsumerSecret)
	token := oauth1.NewToken(p.AccessToken, p.TokenSecret)

	return config.Client(oauth1.NoContext, token)
}
