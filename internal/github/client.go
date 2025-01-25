package github

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/mdobak/go-xerrors"
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

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *Client) User(ctx context.Context) (*User, error) {
	user, _, err := c.c.Users.Get(ctx, "")
	if err != nil {
		return nil, xerrors.New(ErrClient, err)
	}
	return &User{
		Name:  user.GetName(),
		Email: user.GetEmail(),
	}, nil
}
