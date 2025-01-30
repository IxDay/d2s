package github

import (
	"context"

	"github.com/mdobak/go-xerrors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
	oauthConfig *oauth2.Config
	bypassToken string

	ErrMissingRedirect     = xerrors.Message("can't instanciate, missing redirect")
	ErrMissingClientID     = xerrors.Message("can't instanciate, missing client ID")
	ErrMissingClientSecret = xerrors.Message("can't instanciate, missing redirect")
	ErrMissingBypassToken  = xerrors.Message("can't instanciate, missing token")
)

func InitOAuth(redirect, id, secret string) (err error) {
	if oauthConfig != nil {
		panic("oauth config already initialized")
	}
	if redirect == "" {
		err = xerrors.Append(err, ErrMissingRedirect)
	}
	if id == "" {
		err = xerrors.Append(err, ErrMissingClientID)
	}
	if secret == "" {
		err = xerrors.Append(err, ErrMissingClientSecret)
	}
	if err != nil {
		return err
	}
	oauthConfig = &oauth2.Config{
		RedirectURL:  redirect,
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
	return nil
}

func InitBypass(token string) error {
	if token == "" {
		return ErrMissingBypassToken
	}
	bypassToken = token
	return nil
}

func AuthCodeURL(state string) string {
	return oauthConfig.AuthCodeURL(state)
}

func Exchange(ctx context.Context, code string) (string, error) {
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", xerrors.New("failed retrieving token from code", err)
	}
	return token.AccessToken, nil
}
