package github

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/mdobak/go-xerrors"
	"github.com/platipy-io/d2s/types"
)

var ErrClient = xerrors.Message("github API call failed")

var httpClient = github.NewClient(&http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 5,
	},
})

type Client struct {
	c *github.Client
}

func NewClient(token string) *Client {
	return &Client{c: httpClient.WithAuthToken(token)}
}

func UserBypass(ctx context.Context) (*types.User, error) {
	return User(ctx, bypassToken)
}

func User(ctx context.Context, token string) (*types.User, error) {
	user, _, err := NewClient(token).c.Users.Get(ctx, "")
	if err != nil {
		return nil, xerrors.New(ErrClient, err)
	}
	return types.NewUser(user.GetName(), user.GetEmail(), token), nil
}
