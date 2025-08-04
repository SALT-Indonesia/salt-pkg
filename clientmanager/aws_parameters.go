package clientmanager

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type AWSParameters struct {
	signer  *v4.Signer
	Key     string
	Secret  string
	Service string
	Region  string
}

func (p AWSParameters) Credentials(ctx context.Context) (aws.Credentials, error) {
	provider := credentials.NewStaticCredentialsProvider(p.Key, p.Secret, "")
	creds, err := provider.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}
	return creds, nil
}

func (p *AWSParameters) Signer(r *http.Request) error {
	if p.signer == nil {
		p.signer = v4.NewSigner()
	}

	creds, err := p.Credentials(r.Context())
	if err != nil {
		return err
	}

	body := ""
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		body = string(bodyBytes)
	}

	return p.signer.SignHTTP(
		r.Context(),
		creds,
		r,
		body,
		p.Service,
		p.Region,
		time.Now(),
	)
}
