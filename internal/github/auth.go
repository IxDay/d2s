package github

import (
	"context"

	"github.com/mdobak/go-xerrors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var oauthConfig *oauth2.Config

func InitOAuth(redirect, id, secret string) {
	if oauthConfig != nil {
		panic("oauth config already initialized")
	}

	oauthConfig = &oauth2.Config{
		RedirectURL:  redirect,
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
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
